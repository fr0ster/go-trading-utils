package futures_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"

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
		up            *btree.BTree
		down          *btree.BTree

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

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
		callbackRate       float64

		deltaPrice    float64
		deltaQuantity float64

		isArithmetic bool
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
				ActivationPrice(utils.ConvFloat64ToStr(stopPrice, priceRound))
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
			return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1, err)
		} else if apiError.Code == -2021 {
			time.Sleep(3 * time.Second)
			if sideType == futures.SideTypeSell {
				return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1, err)
			} else if sideType == futures.SideTypeBuy {
				return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1, err)
			}
		}
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
	callbackRate float64) (
	order *futures.CreateOrderResponse, err error) {
	return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, 3)
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

func (pp *PairProcessor) GetLeverage() int {
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
	risk, _ := pp.GetPositionRisk()
	return pairs_types.MarginType(strings.ToUpper(risk.MarginType))
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

func (pp *PairProcessor) UserDataEventStart(
	callBack func(event *futures.WsUserDataEvent), eventType ...futures.UserDataEventType) (resetEvent chan error, err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(pp.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Отримуємо ключ для прослуховування подій користувача
	listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
	resetEvent = make(chan error, 1)
	// Ініціалізуємо обробник помилок
	wsErrorHandler := func(err error) {
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
	var stopC chan struct{}
	_, stopC, err = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
	if err != nil {
		return
	}
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-resetEvent:
				// Оновлюємо стан з'єднання для стріму подій користувача з раніше отриманим ключем
				err := pp.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
				if err != nil {
					// Отримуємо новий ключ для прослуховування подій користувача при втраті з'єднання
					listenKey, err = pp.client.NewStartUserStreamService().Do(context.Background())
					if err != nil {
						return
					}
				}
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				// Запускаємо стрім подій користувача
				_, stopC, _ = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
			case <-ticker.C:
				// Оновлюємо стан з'єднання для стріму подій користувача з раніше отриманим ключем
				err := pp.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background())
				if err != nil {
					// Отримуємо новий ключ для прослуховування подій користувача при втраті з'єднання
					listenKey, err = pp.client.NewStartUserStreamService().Do(context.Background())
					if err != nil {
						return
					}
				}
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > pp.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо стрім подій користувача
					_, stopC, _ = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
				}
			}
		}
	}()
	return
}

// Округлення ціни до StepSize знаків після коми
func (pp *PairProcessor) getStepSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func (pp *PairProcessor) getTickSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))))
}

func (pp *PairProcessor) roundPrice(price float64) float64 {
	return utils.RoundToDecimalPlace(price, pp.getTickSizeExp())
}

func (pp *PairProcessor) roundQuantity(quantity float64) float64 {
	return utils.RoundToDecimalPlace(quantity, pp.getStepSizeExp())
}

func (pp *PairProcessor) Steps(begin, end float64, next func(b float64, n int) float64) int {
	var test func(float64, float64) bool
	n := 1
	if begin < end {
		test = func(s, e float64) bool { return s < e }
	} else {
		test = func(s, e float64) bool { return s > e }
	}
	for sum := begin; test(sum, end); sum = next(sum, n) {
		n++
	}
	return n - 1
}

func (pp *PairProcessor) TotalValue(
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
	if pp.isArithmetic {
		n = utils.FindLengthOfArithmeticProgression(P1, P1*(1+deltaPrice), P2)
		Q2 := pp.roundQuantity(utils.FindArithmeticProgressionNthTerm(Q1, Q1*(1+pp.GetDeltaQuantity()), n))
		value = utils.ArithmeticProgressionSum(P1*Q1, P2*Q2, n)
	} else {
		n = utils.FindLengthOfGeometricProgression(P1, P1*(1+deltaPrice), P2)
		Q2 := pp.roundQuantity(utils.FindGeometricProgressionNthTerm(Q1, Q1*(1+pp.GetDeltaQuantity()), n))
		value = utils.GeometricProgressionSum(P1*Q1, P2*Q2, n)
	}
	return
}

func (pp *PairProcessor) recSearch(
	P1 float64,
	low float64,
	high float64,
	P2 float64,
	limit float64,
	minSteps int,
	buffer ...*btree.BTree) (
	value,
	startQuantity float64,
	n int,
	err error) {
	if high-low <= pp.stepSizeDelta {
		value, n = pp.TotalValue(P1, low, P2)
		if value < limit && n >= minSteps {
			return value, low, n, nil
		} else {
			err = fmt.Errorf("can't calculate initial position")
			return
		}
	} else {
		mid := pp.roundQuantity((low + high) / 2)
		value, n := pp.TotalValue(P1, mid, P2)
		if value < limit && n >= minSteps {
			return pp.recSearch(P1, mid, high, P2, limit, minSteps, buffer...)
		} else {
			return pp.recSearch(P1, low, mid, P2, limit, minSteps, buffer...)
		}
	}
}

func (pp *PairProcessor) CalculateInitialPosition(
	minN int,
	buyPrice,
	endPrice float64) (value, quantity float64, n int, err error) {
	var (
		tree *btree.BTree
	)
	low := pp.roundQuantity(pp.notional / buyPrice)
	high := pp.roundQuantity(pp.GetFreeBalance() * float64(pp.GetLeverage()) / buyPrice)
	value, quantity, n, err = pp.recSearch(
		buyPrice,
		low,
		high,
		endPrice,
		pp.GetFreeBalance()*float64(pp.GetLeverage()),
		minN,
		tree)
	if err != nil {
		return
	}
	return
}

func (pp *PairProcessor) InitPositionGrid(
	minN int,
	price float64) (
	valueUp,
	priceUp,
	quantityUp float64,
	stepsUp int,
	valueDown,
	priceDown,
	quantityDown float64,
	stepsDown int,
	err error) {
	priceUp = price * (1 + pp.GetDeltaPrice())
	valueUp, quantityUp, stepsUp, err = pp.CalculateInitialPosition(
		minN,
		priceUp,
		pp.UpBound)
	if err != nil {
		return
	}
	for i := 0; i < stepsUp; i++ {
		if pp.isArithmetic {
			priceN := utils.FindArithmeticProgressionNthTerm(priceUp, priceUp*(1+pp.GetDeltaPrice()), i+1)
			quantityN := utils.FindArithmeticProgressionNthTerm(quantityUp, quantityUp*(1+pp.GetDeltaQuantity()), i+1)
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceN, Quantity: quantityN})
		} else {
			priceN := utils.FindGeometricProgressionNthTerm(priceUp, priceUp*(1+pp.GetDeltaPrice()), i+1)
			quantityN := utils.FindGeometricProgressionNthTerm(quantityUp, quantityUp*(1+pp.GetDeltaQuantity()), i+1)
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceN, Quantity: quantityN})
		}
	}
	priceDown = price * (1 - pp.GetDeltaPrice())
	valueDown, quantityDown, stepsDown, err = pp.CalculateInitialPosition(
		minN,
		priceDown,
		pp.LowBound)
	if err != nil {
		return
	}
	for i := 0; i < stepsUp; i++ {
		if pp.isArithmetic {
			priceN := utils.FindArithmeticProgressionNthTerm(priceDown, priceDown*(1-pp.GetDeltaPrice()), i+1)
			quantityN := utils.FindArithmeticProgressionNthTerm(quantityUp, quantityUp*(1+pp.GetDeltaQuantity()), i+1)
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceN, Quantity: quantityN})
		} else {
			priceN := utils.FindGeometricProgressionNthTerm(priceUp, priceUp*(1-pp.GetDeltaPrice()), i+1)
			quantityN := utils.FindGeometricProgressionNthTerm(quantityUp, quantityUp*(1+pp.GetDeltaQuantity()), i+1)
			pp.up.ReplaceOrInsert(&pair_price_types.PairPrice{Price: priceN, Quantity: quantityN})
		}
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
	if pp.isArithmetic {
		return pp.roundPrice(utils.FindArithmeticProgressionNthTerm(price, price*(1+pp.GetDeltaPrice()), 1))
	} else {
		return pp.roundPrice(utils.FindGeometricProgressionNthTerm(price, price*(1+pp.GetDeltaPrice()), 1))
	}
}

func (pp *PairProcessor) NextPriceDown(price float64) float64 {
	if pp.isArithmetic {
		return pp.roundPrice(utils.FindArithmeticProgressionNthTerm(price, price*(1-pp.GetDeltaPrice()), 1))
	} else {
		return pp.roundPrice(utils.FindGeometricProgressionNthTerm(price, price*(1-pp.GetDeltaPrice()), 1))
	}
}

func (pp *PairProcessor) NextQuantityUp(quantity float64) float64 {
	if pp.isArithmetic {
		return pp.roundQuantity(utils.FindArithmeticProgressionNthTerm(quantity, quantity*(1+pp.GetDeltaQuantity()), 1))
	} else {
		return pp.roundQuantity(utils.FindGeometricProgressionNthTerm(quantity, quantity*(1+pp.GetDeltaQuantity()), 1))
	}
}

func (pp *PairProcessor) NextQuantityDown(quantity float64) float64 {
	if pp.isArithmetic {
		return utils.FindArithmeticProgressionNthTerm(quantity, quantity*(1-pp.GetDeltaQuantity()), 1)
	} else {
		return utils.FindGeometricProgressionNthTerm(quantity, quantity*(1-pp.GetDeltaQuantity()), 1)
	}
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

func NewPairProcessor(
	client *futures.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	leverage int,
	callbackRate float64,
	stop chan struct{},
	isArithmetic bool) (pp *PairProcessor, err error) {
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
		up:            btree.New(2),
		down:          btree.New(2),

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

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

		isArithmetic: isArithmetic,
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
	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit,
		err =
		LimitRead(pp.degree, []string{pp.symbol.Symbol}, client)
	if err != nil {
		return
	}

	return
}
