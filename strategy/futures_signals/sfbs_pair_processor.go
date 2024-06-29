package futures_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"
	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	PairProcessor struct {
		client        *futures.Client
		exchangeInfo  *exchange_types.ExchangeInfo
		symbol        *futures.Symbol
		baseSymbol    string
		targetSymbol  string
		notional      float64
		stepSizeDelta float64
		minSteps      int
		up            *btree.BTree
		down          *btree.BTree

		stop chan struct{}

		pairInfo           *symbol_types.FuturesSymbol
		orderTypes         map[futures.OrderType]bool
		degree             int
		sleepingTime       time.Duration
		timeOut            time.Duration
		limitOnPosition    float64
		limitOnTransaction float64
		UpBound            float64
		LowBound           float64
		leverage           int
		marginType         pairs_types.MarginType
		callbackRate       float64

		deltaPrice    float64
		deltaQuantity float64

		progression             pairs_types.ProgressionType
		GetDelta                progressions.DeltaType
		NthTerm                 progressions.NthTermType
		Sum                     progressions.SumType
		FindNthTerm             progressions.FindNthTermType
		FindLengthOfProgression progressions.FindLengthOfProgressionType
		FindProgressionTthTerm  progressions.FindCubicProgressionTthTermType
	}
)

//  1. Order with type STOP, parameter timeInForce can be sent ( default GTC).
//  2. Order with type TAKE_PROFIT, parameter timeInForce can be sent ( default GTC).
//  3. Condition orders will be triggered when:
//     a) If parameterpriceProtectis sent as true:
//     when price reaches the stopPrice ，the difference rate between "MARK_PRICE" and
//     "CONTRACT_PRICE" cannot be larger than the "triggerProtect" of the symbol
//     "triggerProtect" of a symbol can be got from GET /fapi/v1/exchangeInfo
//     b) STOP, STOP_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     c) TAKE_PROFIT, TAKE_PROFIT_MARKET:
//     BUY: latest price ("MARK_PRICE" or "CONTRACT_PRICE") <= stopPrice
//     SELL: latest price ("MARK_PRICE" or "CONTRACT_PRICE") >= stopPrice
//     d) TRAILING_STOP_MARKET:
//     BUY: the lowest price after order placed <= activationPrice,
//     and the latest price >= the lowest price * (1 + callbackRate)
//     SELL: the highest price after order placed >= activationPrice,
//     and the latest price <= the highest price * (1 - callbackRate)
//  4. For TRAILING_STOP_MARKET, if you got such error code.
//     {"code": -2021, "msg": "Order would immediately trigger."}
//     means that the parameters you send do not meet the following requirements:
//     BUY: activationPrice should be smaller than latest price.
//     SELL: activationPrice should be larger than latest price.
//     If newOrderRespType is sent as RESULT :
//     MARKET order: the final FILLED result of the order will be return directly.
//     LIMIT order with special timeInForce:
//     the final status result of the order(FILLED or EXPIRED)
//     will be returned directly.
//  5. STOP_MARKET, TAKE_PROFIT_MARKET with closePosition=true:
//     Follow the same rules for condition orders.
//     If triggered，close all current long position( if SELL) or current short position( if BUY).
//     Cannot be used with quantity parameter
//     Cannot be used with reduceOnly parameter
//     In Hedge Mode,cannot be used with BUY orders in LONG position side
//     and cannot be used with SELL orders in SHORT position side
//  6. selfTradePreventionMode is only effective when timeInForce set to IOC or GTC or GTD.
//  7. In extreme market conditions,
//     timeInForce GTD order auto cancel time might be delayed comparing to goodTillDate
func (pp *PairProcessor) createOrder(
	orderType futures.OrderType,
	sideType futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity float64,
	closePosition bool,
	price float64,
	stopPrice float64,
	activationPrice float64,
	callbackRate float64,
	times int,
	oldErr ...error) (
	order *futures.CreateOrderResponse, err error) {
	if times == 0 {
		if len(oldErr) == 0 {
			err = fmt.Errorf("can't create order")
		} else {
			err = oldErr[0]
		}
		return
	}
	pp.symbol, err = (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	if _, ok := pp.orderTypes[orderType]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.symbol.Symbol)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderResponseType(futures.NewOrderRespTypeRESULT).
			Symbol(string(futures.SymbolType(pp.symbol.Symbol))).
			Type(orderType).
			Side(sideType)
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == futures.OrderTypeMarket {
		// MARKET	quantity
		service = service.Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
	} else if orderType == futures.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == futures.OrderTypeStop || orderType == futures.OrderTypeTakeProfit {
		// STOP/TAKE_PROFIT	quantity, price, stopPrice
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound)).
			StopPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
	} else if orderType == futures.OrderTypeStopMarket || orderType == futures.OrderTypeTakeProfitMarket {
		// STOP_MARKET/TAKE_PROFIT_MARKET	stopPrice
		service = service.
			StopPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
		if closePosition {
			service = service.ClosePosition(closePosition)
		}
	} else if orderType == futures.OrderTypeTrailingStopMarket {
		// TRAILING_STOP_MARKET	quantity,callbackRate
		service = service.
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			CallbackRate(utils.ConvFloat64ToStr(callbackRate, priceRound))
		if stopPrice != 0 {
			service = service.
				ActivationPrice(utils.ConvFloat64ToStr(activationPrice, priceRound))
		}
	}
	order, err = service.Do(context.Background())
	if err != nil {
		logrus.Errorf("Can't create order: %v", err)
		apiError, _ := utils.ParseAPIError(err)
		if apiError == nil {
			return
		} else if apiError.Code == -1007 {
			time.Sleep(1 * time.Second)
			orders, err := pp.GetOpenOrders()
			if err != nil {
				return nil, err
			}
			for _, order := range orders {
				if order.Symbol == pp.symbol.Symbol &&
					order.Side == sideType &&
					order.Price == utils.ConvFloat64ToStr(price, priceRound) {
					return &futures.CreateOrderResponse{
						Symbol:                  order.Symbol,
						OrderID:                 order.OrderID,
						ClientOrderID:           order.ClientOrderID,
						Price:                   order.Price,
						OrigQuantity:            order.OrigQuantity,
						ExecutedQuantity:        order.ExecutedQuantity,
						CumQuote:                order.CumQuote,
						ReduceOnly:              order.ReduceOnly,
						Status:                  order.Status,
						StopPrice:               order.StopPrice,
						TimeInForce:             order.TimeInForce,
						Type:                    order.Type,
						Side:                    order.Side,
						UpdateTime:              order.UpdateTime,
						WorkingType:             order.WorkingType,
						ActivatePrice:           order.ActivatePrice,
						PriceRate:               order.PriceRate,
						AvgPrice:                order.AvgPrice,
						PositionSide:            order.PositionSide,
						ClosePosition:           order.ClosePosition,
						PriceProtect:            order.PriceProtect,
						PriceMatch:              order.PriceMatch,
						SelfTradePreventionMode: order.SelfTradePreventionMode,
						GoodTillDate:            order.GoodTillDate,
						CumQty:                  order.CumQuantity,
						OrigType:                order.OrigType,
					}, nil
				}
			}
			// На наступних кодах помилок можна спробувати ще раз
		} else if apiError.Code == -1008 || apiError.Code == -5028 {
			time.Sleep(3 * time.Second)
			return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, activationPrice, callbackRate, times-1, err)
		}
		return
	}
	return
}
func (pp *PairProcessor) CreateOrder(
	orderType futures.OrderType,
	sideType futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity float64,
	closePosition bool,
	price float64,
	stopPrice float64,
	activationPrice float64,
	callbackRate float64) (
	order *futures.CreateOrderResponse, err error) {
	return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, activationPrice, callbackRate, 3)
}

func (pp *PairProcessor) GetOpenOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *futures.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.symbol.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *futures.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.symbol.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (err error) {
	return pp.client.NewCancelAllOpenOrdersService().Symbol(pp.symbol.Symbol).Do(context.Background())
}

func (pp *PairProcessor) getPositionRisk(times int) (risks []*futures.PositionRisk, err error) {
	if times == 0 {
		return
	}
	risks, err = pp.client.NewGetPositionRiskService().Symbol(pp.pairInfo.GetSymbol()).Do(context.Background())
	if err != nil {
		errApi, _ := utils.ParseAPIError(err)
		if errApi != nil && errApi.Code == -1021 {
			time.Sleep(3 * time.Second)
			return pp.getPositionRisk(times - 1)
		}
	}
	return
}

func (pp *PairProcessor) GetPositionRisk() (risks *futures.PositionRisk, err error) {
	risk, err := pp.getPositionRisk(3)
	if err != nil {
		return nil, err
	} else if len(risk) == 0 {
		return nil, fmt.Errorf("can't get position risk for symbol %s", pp.symbol.Symbol)
	} else {
		return risk[0], nil
	}
}

func (pp *PairProcessor) GetLiquidationDistance(price float64) (distance float64) {
	risk, _ := pp.GetPositionRisk()
	return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
}

func (pp *PairProcessor) GetNotional() float64 {
	return pp.notional
}

func (pp *PairProcessor) GetLeverage() int {
	if pp.leverage == 0 {
		risk, _ := pp.GetPositionRisk()
		pp.leverage = int(utils.ConvStrToFloat64(risk.Leverage))
	}
	return pp.leverage
}

func (pp *PairProcessor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	return pp.client.NewChangeLeverageService().Symbol(pp.symbol.Symbol).Leverage(leverage).Do(context.Background())
}

func (pp *PairProcessor) GetCallbackRate() float64 {
	return pp.callbackRate
}

func (pp *PairProcessor) SetCallbackRate(callbackRate float64) {
	pp.callbackRate = callbackRate
}

func (pp *PairProcessor) GetDeltaPrice() float64 {
	return pp.deltaPrice
}

func (pp *PairProcessor) SetDeltaPrice(deltaPrice float64) {
	pp.deltaPrice = deltaPrice
}

func (pp *PairProcessor) GetDeltaQuantity() float64 {
	return pp.deltaQuantity
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) GetMarginType() pairs_types.MarginType {
	if pp.marginType == "" {
		risk, _ := pp.GetPositionRisk()
		pp.marginType = pairs_types.MarginType(risk.MarginType)
	}
	return pp.marginType
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) SetMarginType(marginType pairs_types.MarginType) (err error) {
	return pp.client.
		NewChangeMarginTypeService().
		Symbol(pp.symbol.Symbol).
		MarginType(futures.MarginType(marginType)).
		Do(context.Background())
}

func (pp *PairProcessor) GetPositionMargin() (margin float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	return
}

func (pp *PairProcessor) SetPositionMargin(amountMargin float64, typeMargin int) (err error) {
	return pp.client.NewUpdatePositionMarginService().
		Symbol(pp.symbol.Symbol).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(amountMargin)).Do(context.Background())
}

func (pp *PairProcessor) GetSymbol() *symbol_types.FuturesSymbol {
	// Ініціалізуємо інформацію про пару
	return pp.pairInfo
}

func (pp *PairProcessor) GetFuturesSymbol() (*futures.Symbol, error) {
	return pp.symbol, nil
}

// Округлення ціни до StepSize знаків після коми
func (pp *PairProcessor) GetStepSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func (pp *PairProcessor) GetTickSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))))
}

func (pp *PairProcessor) GetAccount() (account *futures.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetPair() string {
	return pp.symbol.Symbol
}

func (pp *PairProcessor) GetBaseAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.baseSymbol {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.baseSymbol)
}

func (pp *PairProcessor) GetTargetAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.targetSymbol {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.targetSymbol)
}

func (pp *PairProcessor) GetBaseBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.WalletBalance) // Convert string to float64
	return
}

func (pp *PairProcessor) GetTargetBalance() (balance float64, err error) {
	asset, err := pp.GetTargetAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	return
}

func (pp *PairProcessor) GetFreeBalance() (balance float64) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return 0
	}
	balance = utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	if balance > pp.limitOnPosition {
		balance = pp.limitOnPosition
	}
	return
}

func (pp *PairProcessor) GetLimitOnTransaction() (limit float64) {
	return pp.limitOnTransaction * pp.GetFreeBalance()
}

func (pp *PairProcessor) GetUpBound() float64 {
	return pp.UpBound
}

func (pp *PairProcessor) GetLowBound() float64 {
	return pp.LowBound
}

func (pp *PairProcessor) GetLockedBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	return
}

func (pp *PairProcessor) Debug(fl, id string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		orders, _ := pp.GetOpenOrders()
		logrus.Debugf("%s %s %s:", fl, id, pp.symbol.Symbol)
		for _, order := range orders {
			logrus.Debugf(" Open Order %v on price %v OrderSide %v Status %s", order.OrderID, order.Price, order.Side, order.Status)
		}
	}
}

func (pp *PairProcessor) StartStream(handler futures.WsUserDataHandler, errHandler futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// Отримуємо новий або той же самий ключ для прослуховування подій користувача при втраті з'єднання
	listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	// Запускаємо стрім подій користувача
	doneC, stopC, err = futures.WsUserDataServe(listenKey, handler, errHandler)
	return
}

func (pp *PairProcessor) UserDataEventStart(
	callBack func(event *futures.WsUserDataEvent),
	eventType ...futures.UserDataEventType) (resetEvent chan error, err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(pp.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
	resetEvent = make(chan error, 1)
	// Ініціалізуємо обробник помилок
	wsErrorHandler := func(err error) {
		logrus.Errorf("Future wsErrorHandler error: %v", err)
		resetEvent <- err
	}
	// Ініціалізуємо обробник подій
	eventMap := make(map[futures.UserDataEventType]bool)
	for _, event := range eventType {
		eventMap[event] = true
	}
	wsHandler := func(event *futures.WsUserDataEvent) {
		if len(eventType) == 0 || eventMap[event.Event] {
			callBack(event)
		}
	}
	// Запускаємо стрім подій користувача
	doneC, stopC, err := pp.StartStream(wsHandler, wsErrorHandler)
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-doneC:
				// Запускаємо новий стрім подій користувача
				doneC, stopC, err = pp.StartStream(wsHandler, wsErrorHandler)
				return
			case <-resetEvent:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > pp.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
				}
			}
		}
	}()
	return
}

func (pp *PairProcessor) RoundPrice(price float64) float64 {
	return utils.RoundToDecimalPlace(price, pp.GetTickSizeExp())
}

func (pp *PairProcessor) RoundQuantity(quantity float64) float64 {
	return utils.RoundToDecimalPlace(quantity, pp.GetStepSizeExp())
}

func (pp *PairProcessor) CalcValueForQuantity(
	P1 float64,
	Q1 float64,
	P2 float64) (
	value float64,
	n int) {
	var (
		deltaPrice float64
	)
	if P1 < P2 {
		deltaPrice = pp.GetDeltaPrice()
	} else {
		deltaPrice = -pp.GetDeltaPrice()
	}
	n = pp.FindLengthOfProgression(P1, P1*(1+deltaPrice), P2)
	delta := pp.GetDelta(P1*Q1, P1*(1+deltaPrice)*Q1*(1+pp.GetDeltaQuantity()))
	value = pp.Sum(P1*Q1, delta, n)
	return
}

func (pp *PairProcessor) CalculateInitialPosition(
	minN int,
	buyPrice,
	endPrice float64) (value, quantity float64, n int, err error) {
	low := pp.RoundQuantity(pp.notional / buyPrice)
	high := pp.RoundQuantity(pp.limitOnPosition * float64(pp.leverage) / buyPrice)

	for pp.RoundQuantity(high-low) > pp.stepSizeDelta {
		mid := pp.RoundQuantity((low + high) / 2)
		value, n = pp.CalcValueForQuantity(buyPrice, mid, endPrice)
		if value <= pp.limitOnPosition*float64(pp.leverage) && n >= minN {
			low = mid
		} else {
			high = mid
		}
	}

	value, n = pp.CalcValueForQuantity(buyPrice, high, endPrice)
	if value < pp.limitOnPosition*float64(pp.leverage) && n >= minN {
		quantity = pp.RoundQuantity(high)
		return
	}
	value, n = pp.CalcValueForQuantity(buyPrice, low, endPrice)
	if value < pp.limitOnPosition*float64(pp.leverage) && n >= minN {
		quantity = pp.RoundQuantity(low)
		return
	}

	err = fmt.Errorf("can't calculate initial position")
	return
}

func (pp *PairProcessor) InitPositionGrid(
	minN int,
	price float64) (
	valueUp,
	startQuantityUp float64,
	stepsUp int,
	valueDown,
	startQuantityDown float64,
	stepsDown int,
	err error) {
	var (
		priceUp      float64
		quantityUp   float64
		priceDown    float64
		quantityDown float64
	)
	valueUp, startQuantityUp, stepsUp, err = pp.CalculateInitialPosition(
		minN,
		price,
		pp.UpBound)
	if err != nil {
		return
	}
	priceUp = price * (1 + pp.GetDeltaPrice())
	quantityUp = startQuantityUp
	pp.up.Clear(false)
	for i := 2; i < stepsUp; i++ {
		pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceUp, Quantity: quantityUp})
		priceUp = pp.RoundPrice(pp.FindNthTerm(priceUp, priceUp*(1+pp.GetDeltaPrice()), i+1))
		quantityUp = pp.RoundQuantity(pp.FindNthTerm(quantityUp, quantityUp*(1+pp.GetDeltaQuantity()), i+1))
	}
	valueDown, startQuantityDown, stepsDown, err = pp.CalculateInitialPosition(
		minN,
		price,
		pp.LowBound)
	if err != nil {
		return
	}
	priceDown = price * (1 - pp.GetDeltaPrice())
	quantityDown = startQuantityDown
	pp.down.Clear(false)
	for i := 2; i < stepsUp; i++ {
		pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceDown, Quantity: quantityDown})
		priceDown = pp.FindNthTerm(priceDown, priceDown*(1-pp.GetDeltaPrice()), i+1)
		quantityDown = pp.FindNthTerm(quantityDown, quantityDown*(1+pp.GetDeltaQuantity()), i+1)
	}
	if quantityUp*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v", pp.notional, quantityUp*price)
	}
	if quantityDown*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v", pp.notional, quantityDown*price)
	}
	return

}

func (pp *PairProcessor) NextPriceUp(price float64) float64 {
	return pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextPriceDown(price float64) float64 {
	return pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
}

func (pp *PairProcessor) NextQuantityUp(quantity float64) float64 {
	return pp.RoundQuantity(quantity * (1 + pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextQuantityDown(quantity float64) float64 {
	return pp.RoundQuantity(quantity * (1 - pp.GetDeltaQuantity()))
}

func (pp *PairProcessor) NextUp(currentPrice, currentQuantity float64) (price, quantity float64, err error) {
	if val := pp.up.Min(); val != nil {
		pair := val.(*pair_price_types.PairPrice)
		pp.up.Delete(val)
		pp.down.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: currentQuantity})
		return pair.Price, pair.Quantity, nil
	} else {
		return 0, 0, fmt.Errorf("can't get next up price")
	}
}

func (pp *PairProcessor) NextDown(currentPrice, currentQuantity float64) (price, quantity float64, err error) {
	if val := pp.down.Max(); val != nil {
		pair := val.(*pair_price_types.PairPrice)
		pp.down.Delete(val)
		pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: currentPrice, Quantity: currentQuantity})
		return pair.Price, pair.Quantity, nil
	} else {
		return 0, 0, fmt.Errorf("can't get next down price")
	}
}

func (pp *PairProcessor) LimitRead() (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	err error) {
	exchangeInfo := exchange_types.New()
	futures_exchange_info.RestrictedInit(exchangeInfo, degree, []string{pp.symbol.Symbol}, pp.client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	if minuteRawRequestLimit == nil {
		err = fmt.Errorf("minute raw request limit is not found")
		return
	}
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}

func (pp *PairProcessor) GetCurrentPrice() (float64, error) {
	price, err := pp.client.NewListPricesService().Symbol(pp.symbol.Symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func (pp *PairProcessor) GetPrices(
	price float64,
	upPositionNewOrderType futures.OrderType,
	downPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	isDynamic bool) (
	priceUp,
	quantityUp,
	priceDown,
	quantityDown float64,
	upOrderType,
	downOrderType futures.OrderType, err error) {
	var (
		risk *futures.PositionRisk
	)
	risk, err = pp.GetPositionRisk()
	if err != nil {
		return
	}
	priceUp = pp.RoundPrice(price * (1 + pp.GetDeltaPrice()))
	priceDown = pp.RoundPrice(price * (1 - pp.GetDeltaPrice()))
	if isDynamic {
		_, quantityUp, _, err = pp.CalculateInitialPosition(pp.minSteps, priceUp, pp.UpBound)
		if err != nil {
			err = fmt.Errorf("can't calculate initial position for price up %v", priceUp)
			printError()
			return
		}
		_, quantityDown, _, err = pp.CalculateInitialPosition(pp.minSteps, priceDown, pp.LowBound)
		if err != nil {
			err = fmt.Errorf("can't calculate initial position for price down %v", priceDown)
			printError()
			return
		}
	} else {
		quantityUp = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceUp)
		quantityDown = pp.RoundQuantity(pp.GetLimitOnTransaction() * float64(pp.GetLeverage()) / priceDown)
	}
	upOrderType = upPositionNewOrderType
	downOrderType = downPositionNewOrderType
	if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		positionPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
		if positionPrice == 0 {
			positionPrice = utils.ConvStrToFloat64(risk.EntryPrice)
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			priceDown = pp.NextPriceDown(math.Min(positionPrice, price))
			quantityDown = math.Max(-utils.ConvStrToFloat64(risk.PositionAmt), quantityDown)
			upOrderType = shortPositionIncOrderType
			downOrderType = shortPositionDecOrderType
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			priceUp = pp.NextPriceUp(math.Max(positionPrice, price))
			quantityUp = math.Max(utils.ConvStrToFloat64(risk.PositionAmt), quantityUp)
			upOrderType = longPositionIncOrderType
			downOrderType = longPositionDecOrderType
		}
	}
	return
}

func NewPairProcessor(
	client *futures.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	callbackRate float64,
	progression pairs_types.ProgressionType,
	stop chan struct{}) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = futures_exchange_info.Init(exchangeInfo, 3, client)
	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:        client,
		exchangeInfo:  exchangeInfo,
		symbol:        nil,
		baseSymbol:    "",
		notional:      0,
		stepSizeDelta: 0,
		minSteps:      minSteps,
		up:            btree.New(2),
		down:          btree.New(2),

		stop: stop,

		pairInfo:           nil,
		orderTypes:         nil,
		degree:             3,
		sleepingTime:       1 * time.Second,
		timeOut:            1 * time.Hour,
		limitOnPosition:    limitOnPosition,
		limitOnTransaction: limitOnTransaction,
		UpBound:            UpBound,
		LowBound:           LowBound,
		leverage:           leverage,
		callbackRate:       callbackRate,

		deltaPrice:    deltaPrice,
		deltaQuantity: deltaQuantity,

		progression: progression,
	}

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: symbol}).(*symbol_types.FuturesSymbol)

	// Ініціалізуємо типи ордерів
	pp.orderTypes = make(map[futures.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderType {
		pp.orderTypes[orderType] = true
	}

	// Буферизуємо інформацію про символ
	pp.symbol, err = pp.GetSymbol().GetFuturesSymbol()
	if err != nil {
		return
	}
	pp.baseSymbol = pp.symbol.QuoteAsset
	pp.targetSymbol = pp.symbol.BaseAsset
	pp.notional = utils.ConvStrToFloat64(pp.symbol.MinNotionalFilter().Notional)
	pp.stepSizeDelta = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)

	if pp.progression == pairs_types.ArithmeticProgression {
		pp.NthTerm = progressions.ArithmeticProgressionNthTerm
		pp.Sum = progressions.ArithmeticProgressionSum
		pp.FindNthTerm = progressions.FindArithmeticProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfArithmeticProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 - P1 }
		pp.FindProgressionTthTerm = progressions.FindArithmeticProgressionTthTerm
	} else if pp.progression == pairs_types.GeometricProgression {
		pp.NthTerm = progressions.GeometricProgressionNthTerm
		pp.Sum = progressions.GeometricProgressionSum
		pp.FindNthTerm = progressions.FindGeometricProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfGeometricProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 / P1 }
		pp.FindProgressionTthTerm = progressions.FindGeometricProgressionTthTerm
	} else if pp.progression == pairs_types.CubicProgression {
		pp.NthTerm = progressions.CubicProgressionNthTerm
		pp.Sum = progressions.CubicProgressionSum
		pp.FindNthTerm = progressions.FindCubicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfCubicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return math.Pow(P2/P1, 1.0/3) }
		pp.FindProgressionTthTerm = progressions.FindCubicProgressionTthTerm
	} else if pp.progression == pairs_types.CubicRootProgression {
		pp.NthTerm = progressions.CubicRootProgressionNthTerm
		pp.Sum = progressions.CubicRootProgressionSum
		pp.FindNthTerm = progressions.FindCubicRootProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfCubicRootProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return math.Cbrt(P2 / P1) }
		pp.FindProgressionTthTerm = progressions.FindCubicRootProgressionTthTerm
	} else if pp.progression == pairs_types.QuadraticProgression {
		pp.NthTerm = progressions.QuadraticProgressionNthTerm
		pp.Sum = progressions.QuadraticProgressionSum
		pp.FindNthTerm = progressions.FindQuadraticProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfQuadraticProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return (P2 - P1) / 1 }
		pp.FindProgressionTthTerm = progressions.FindQuadraticProgressionTthTerm
	} else if pp.progression == pairs_types.ExponentialProgression {
		pp.NthTerm = progressions.ExponentialProgressionNthTerm
		pp.Sum = progressions.ExponentialProgressionSum
		pp.FindNthTerm = progressions.FindExponentialProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfExponentialProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 / P1 }
		pp.FindProgressionTthTerm = progressions.FindExponentialProgressionTthTerm
	} else if pp.progression == pairs_types.LogarithmicProgression {
		pp.NthTerm = progressions.LogarithmicProgressionNthTerm
		pp.Sum = progressions.LogarithmicProgressionSum
		pp.FindNthTerm = progressions.FindLogarithmicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfLogarithmicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return (P2 - P1) / math.Log(2) }
		pp.FindProgressionTthTerm = progressions.FindLogarithmicProgressionTthTerm
	} else if pp.progression == pairs_types.HarmonicProgression {
		pp.NthTerm = progressions.HarmonicProgressionNthTerm
		pp.Sum = progressions.HarmonicProgressionSum
		pp.FindNthTerm = progressions.FindHarmonicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfHarmonicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return 1/P2 - 1/P1 }
		pp.FindProgressionTthTerm = progressions.FindHarmonicProgressionTthTerm
	} else {
		err = fmt.Errorf("progression type %v is not supported", pp.progression)
		return
	}
	if pp.GetMarginType() != marginType {
		_ = pp.SetMarginType(marginType)
	}
	if pp.GetLeverage() != leverage {
		_, _ = pp.SetLeverage(leverage)
	}

	return
}
