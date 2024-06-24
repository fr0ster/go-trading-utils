package futures_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

type (
	nextPriceFunc    func(float64, int) float64
	nextQuantityFunc func(float64, int) float64
	testFunc         func(float64, float64) bool

	PairProcessor struct {
		client        *futures.Client
		pair          *pairs_types.Pairs
		exchangeInfo  *exchange_types.ExchangeInfo
		symbol        *futures.Symbol
		notional      float64
		stepSizeDelta float64
		up            *btree.BTree
		down          *btree.BTree

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

		stop chan struct{}

		pairInfo     *symbol_types.FuturesSymbol
		orderTypes   map[futures.OrderType]bool
		degree       int
		debug        bool
		sleepingTime time.Duration
		timeOut      time.Duration
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
	times int) (
	order *futures.CreateOrderResponse, err error) {
	if times == 0 {
		err = fmt.Errorf("can't create order")
		return
	}
	symbol, err := (*pp.pairInfo).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	if _, ok := pp.orderTypes[orderType]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.pair.GetPair())
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderResponseType(futures.NewOrderRespTypeRESULT).
			Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
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
		apiError, _ := utils.ParseAPIError(err)
		if apiError == nil {
			return
		}
		if apiError.Code == -1007 {
			time.Sleep(1 * time.Second)
			orders, err := pp.GetOpenOrders()
			if err != nil {
				return nil, err
			}
			for _, order := range orders {
				if order.Symbol == pp.GetPair().GetPair() && order.Side == sideType && order.Price == utils.ConvFloat64ToStr(price, priceRound) {
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
			return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1)
		} else if apiError.Code == -2021 {
			time.Sleep(3 * time.Second)
			if sideType == futures.SideTypeSell {
				return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1)
			} else if sideType == futures.SideTypeBuy {
				return pp.createOrder(orderType, sideType, timeInForce, quantity, closePosition, price, stopPrice, callbackRate, times-1)
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

func (pp *PairProcessor) ClosePosition(side futures.SideType, price float64, exp int) (res *futures.CreateOrderResponse, err error) {
	return pp.client.NewCreateOrderService().
		Symbol(string(futures.SymbolType(pp.pair.GetPair()))).
		Type(futures.OrderTypeMarket).
		Side(side).
		Price(utils.ConvFloat64ToStr(price, exp)).
		StopPrice(utils.ConvFloat64ToStr(price, exp)).
		ClosePosition(true).
		Do(context.Background())
}

func (pp *PairProcessor) LimitUpdaterStream() {
	var err error
	go func() {
		for {
			select {
			case <-time.After(pp.updateTime):
				pp.updateTime,
					pp.minuteOrderLimit,
					pp.dayOrderLimit,
					pp.minuteRawRequestLimit,
					err = LimitRead(pp.degree, []string{pp.pair.GetPair()}, pp.client)
				if err != nil {
					logrus.Errorf("Can't update limits: %v", err)
					close(pp.stop)
					return
				}
			case <-pp.stop:
				return
			}
		}
	}()

	// Перевіряємо чи не вийшли за ліміти на запити та ордери
	go func() {
		for {
			select {
			case <-pp.stop:
				return
			default:
			}
			time.Sleep(pp.updateTime)
		}
	}()
}

func (pp *PairProcessor) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairProcessor) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func (pp *PairProcessor) CheckOrderType(orderType futures.OrderType) bool {
	_, ok := pp.orderTypes[orderType]
	return ok
}

func (pp *PairProcessor) GetOpenOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*futures.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *futures.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *futures.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.pair.GetPair()).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (err error) {
	return pp.client.NewCancelAllOpenOrdersService().Symbol(pp.pair.GetPair()).Do(context.Background())
}

// func (pp *PairProcessor) GetUserDataEvent() chan *futures.WsUserDataEvent {
// 	return pp.userDataEvent
// }

// func (pp *PairProcessor) GetOrderStatusEvent() chan *futures.WsUserDataEvent {
// 	return pp.orderStatusEvent
// }

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
		return nil, fmt.Errorf("can't get position risk for symbol %s", pp.pair.GetPair())
	} else {
		return risk[0], nil
	}
}

func (pp *PairProcessor) GetLeverage() int {
	risk, _ := pp.GetPositionRisk()
	leverage, _ := strconv.Atoi(risk.Leverage) // Convert string to int
	return leverage
}

func (pp *PairProcessor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	return pp.client.NewChangeLeverageService().Symbol(pp.pair.GetPair()).Leverage(leverage).Do(context.Background())
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
		Symbol(pp.pair.GetPair()).
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
		Symbol(pp.pair.GetPair()).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(amountMargin)).Do(context.Background())
}

func (pp *PairProcessor) GetPair() *pairs_types.Pairs {
	return pp.pair
}

func (pp *PairProcessor) GetSymbol() *symbol_types.FuturesSymbol {
	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: pp.pair.GetPair()}).(*symbol_types.FuturesSymbol)
	return pp.pairInfo
}

func (pp *PairProcessor) GetAccount() (account *futures.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetBaseAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.pair.GetBaseSymbol() {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.pair.GetBaseSymbol())
}

func (pp *PairProcessor) GetTargetAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.pair.GetTargetSymbol() {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.pair.GetTargetSymbol())
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

func (pp *PairProcessor) GetFreeBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	return
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
		logrus.Debugf("%s %s %s:", fl, id, pp.pair.GetPair())
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

func (pp *PairProcessor) GetPriceDelta(price float64, n int) float64 {
	return pp.roundPrice(price * (math.Pow(1+pp.pair.GetSellDelta(), float64(n))))
}

func (pp *PairProcessor) GetQuantityDelta(quantity float64, n int) float64 {
	return pp.roundQuantity(quantity * (math.Pow(1+pp.pair.GetSellDeltaQuantity(), float64(n))))
}

func (pp *PairProcessor) TotalValue(
	P1 float64,
	Q1 float64,
	P2 float64,
	limit float64,
	minSteps int,
	test testFunc,
	nextPriceFunc nextPriceFunc,
	nextQuantityFunc nextQuantityFunc,
	buffer ...*btree.BTree) (
	value,
	quantity,
	lastPrice,
	startQuantity float64,
	n int,
	err error) {
	n = 1
	lastPrice = pp.roundPrice(nextPriceFunc(P1, n))
	if P1 == P2 || lastPrice == P2 {
		return P1 * Q1, Q1, P1, Q1, 1, fmt.Errorf("P1 %v == P2 %v, can't calculate", P1, P2)
	}
	if P1*Q1 >= limit || lastPrice*Q1 >= limit {
		return P1 * Q1, Q1, P1, Q1, 1, fmt.Errorf("P1*Q1 %v >= limit %v, can't calculate", P1*Q1, limit)
	}
	lastQuantity := Q1
	if buffer != nil {
		buffer[0].Clear(true)
		buffer[0].ReplaceOrInsert(&pair_price_types.PairPrice{Price: lastPrice, Quantity: lastQuantity})
	}
	for {
		quantity += lastQuantity
		value += lastPrice * lastQuantity
		nextQuantity := pp.roundQuantity(nextQuantityFunc(lastQuantity, n))
		nextPrice := pp.roundPrice(nextPriceFunc(lastPrice, n))
		if test(nextPrice, P2) && value+nextPrice*nextQuantity < limit {
			lastQuantity = nextQuantity
			lastPrice = nextPrice
			n++
			if buffer != nil {
				buffer[0].ReplaceOrInsert(&pair_price_types.PairPrice{Price: lastPrice, Quantity: lastQuantity})
			}
			continue
		} else {
			break
		}
	}
	if n < minSteps {
		return value, pp.roundQuantity(quantity), lastPrice, pp.roundQuantity(Q1), n, fmt.Errorf("n == 0, can't calculate")
	}
	return value, pp.roundQuantity(quantity), lastPrice, pp.roundQuantity(Q1), n, nil
}

func (pp *PairProcessor) CalculateInitialPosition(
	minN int,
	buyPrice,
	endPrice,
	priceDeltaPercent,
	quantityDeltaPercent float64) (value, price, quantity float64, n int, err error) {
	var (
		test         testFunc
		nextPrice    nextPriceFunc
		nextQuantity nextQuantityFunc
	)
	logrus.Debugf("Calculate initial position: budget %v, minN %v, leverage %v, minValue %v, buyPrice %v, endPrice %v, priceDeltaPercent %v, quantityDeltaPercent %v",
		pp.pair.GetCurrentPositionBalance(),
		minN,
		pp.GetLeverage(),
		pp.notional,
		buyPrice,
		endPrice,
		priceDeltaPercent,
		quantityDeltaPercent)
	var (
		tree  *btree.BTree
		testQ float64
	)
	if buyPrice < endPrice {
		test = func(s, e float64) bool { return s < e }
		nextPrice = func(s float64, n int) float64 { return pp.roundPrice(s * math.Pow(1+priceDeltaPercent, float64(2))) }
		tree = pp.up
	} else {
		test = func(s, e float64) bool { return s > e }
		nextPrice = func(s float64, n int) float64 { return pp.roundPrice(s * math.Pow(1-priceDeltaPercent, float64(2))) }
		tree = pp.down
	}
	nextQuantity = func(s float64, n int) float64 {
		return pp.roundQuantity(s * (math.Pow(1+quantityDeltaPercent, float64(2))))
	}
	low := pp.roundQuantity(pp.notional / buyPrice)
	high := pp.roundQuantity(pp.pair.GetCurrentPositionBalance() * float64(pp.GetLeverage()) / buyPrice)
	for testQ = high; testQ >= low; testQ -= pp.stepSizeDelta {
		value, _, price, quantity, n, err = pp.TotalValue(
			buyPrice,
			pp.roundQuantity(testQ),
			endPrice,
			pp.pair.GetCurrentPositionBalance()*float64(pp.pair.GetLeverage()),
			minN,
			test,
			nextPrice,
			nextQuantity,
			tree)
		if err == nil && n >= minN {
			break
		}
	}
	logrus.Debugf("Calculate initial position: value %v, price %v, quantity %v, n %v, err %v",
		value, price, quantity, n, err)
	return
}

func (pp *PairProcessor) InitPositionGrid(
	minN int,
	price float64) (
	priceUp,
	quantityUp,
	priceDown,
	quantityDown float64, err error) {
	_, priceUp, quantityUp, _, err = pp.CalculateInitialPosition(
		minN,
		price,
		pp.pair.GetUpBound(),
		pp.pair.GetSellDelta(),
		pp.pair.GetBuyDeltaQuantity())
	if err != nil {
		return
	}
	_, priceDown, quantityDown, _, err = pp.CalculateInitialPosition(
		minN,
		price,
		pp.pair.GetLowBound(),
		pp.pair.GetSellDelta(),
		pp.pair.GetBuyDeltaQuantity())
	if err != nil {
		return
	}
	if quantityUp*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone up: %v but can buy only for %v", pp.notional, quantityUp*price)
	}
	if quantityDown*price < pp.notional {
		err = fmt.Errorf("we need more money for position if price gone down: %v but can buy only for %v", pp.notional, quantityDown*price)
	}
	if val := pp.up.Min(); val != nil {
		priceUp = val.(*pair_price_types.PairPrice).Price
		pp.up.Delete(val)
	} else {
		err = fmt.Errorf("can't get price up")
	}
	if val := pp.down.Max(); val != nil {
		priceDown = val.(*pair_price_types.PairPrice).Price
		pp.down.Delete(val)
	} else {
		err = fmt.Errorf("can't get price down")
	}
	return

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

func NewPairProcessor(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	stop chan struct{},
	debug bool) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = futures_exchange_info.Init(exchangeInfo, 3, client)

	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:        client,
		pair:          pair,
		exchangeInfo:  exchangeInfo,
		symbol:        nil,
		notional:      0,
		stepSizeDelta: 0,
		up:            btree.New(2),
		down:          btree.New(2),

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		stop: stop,

		pairInfo:     nil,
		orderTypes:   nil,
		degree:       3,
		debug:        debug,
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit,
		err =
		LimitRead(pp.degree, []string{pp.pair.GetPair()}, client)
	if err != nil {
		return
	}

	// Буферизуємо інформацію про символ
	pp.symbol, err = pp.GetSymbol().GetFuturesSymbol()
	if err != nil {
		return
	}
	pp.notional = utils.ConvStrToFloat64(pp.symbol.MinNotionalFilter().Notional)
	pp.stepSizeDelta = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: pair.GetPair()}).(*symbol_types.FuturesSymbol)

	// Ініціалізуємо типи ордерів
	pp.orderTypes = make(map[futures.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderType {
		pp.orderTypes[orderType] = true
	}

	return
}
