package spot_signals

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	nextPriceFunc    func(float64, int) float64
	nextQuantityFunc func(float64, int) float64
	testFunc         func(float64, float64) bool
	Functions        struct {
		NextPriceUp      nextPriceFunc
		NextPriceDown    nextPriceFunc
		NextQuantityUp   nextQuantityFunc
		NextQuantityDown nextQuantityFunc
		TestUp           testFunc
		TestDown         testFunc
	}
	PairProcessor struct {
		client        *binance.Client
		exchangeInfo  *exchange_types.ExchangeInfo
		symbol        *futures.Symbol
		baseSymbol    string
		targetSymbol  string
		notional      float64
		stepSizeDelta float64

		updateTime            time.Duration
		minuteOrderLimit      *exchange_types.RateLimits
		dayOrderLimit         *exchange_types.RateLimits
		minuteRawRequestLimit *exchange_types.RateLimits

		stop chan struct{}

		pairInfo           *symbol_types.SpotSymbol
		orderTypes         map[string]bool
		degree             int
		debug              bool
		sleepingTime       time.Duration
		timeOut            time.Duration
		limitOnPosition    float64
		limitOnTransaction float64
		UpBound            float64
		LowBound           float64
		callbackRate       float64

		deltaPrice    float64
		deltaQuantity float64

		testUp           testFunc
		testDown         testFunc
		nextPriceUp      nextPriceFunc
		nextPriceDown    nextPriceFunc
		nextQuantityUp   nextQuantityFunc
		nextQuantityDown nextQuantityFunc
	}
)

//  1. LIMIT_MAKER are LIMIT orders that will be rejected if they would immediately match and trade as a taker.
//  2. STOP_LOSS and TAKE_PROFIT will execute a MARKET order when the stopPrice is reached.
//     Any LIMIT or LIMIT_MAKER type order can be made an iceberg order by sending an icebergQty.
//     Any order with an icebergQty MUST have timeInForce set to GTC.
//  3. MARKET orders using the quantity field specifies the amount of the base asset the user wants to buy or sell at the market price.
//     For example, sending a MARKET order on BTCUSDT will specify how much BTC the user is buying or selling.
//  4. MARKET orders using quoteOrderQty specifies the amount the user wants to spend (when buying) or receive (when selling) the quote asset;
//     the correct quantity will be determined based on the market liquidity and quoteOrderQty.
//     Using BTCUSDT as an example:
//     On the BUY side, the order will buy as many BTC as quoteOrderQty USDT can.
//     On the SELL side, the order will sell as much BTC needed to receive quoteOrderQty USDT.
//  5. MARKET orders using quoteOrderQty will not break LOT_SIZE filter rules; the order will execute a quantity that will have the notional value as close as possible to quoteOrderQty.
//     same newClientOrderId can be accepted only when the previous one is filled, otherwise the order will be rejected.
//  6. For STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT and TAKE_PROFIT orders, trailingDelta can be combined with stopPrice.
//
//  7. Trigger order price rules against market price for both MARKET and LIMIT versions:
//     Price above market price: STOP_LOSS BUY, TAKE_PROFIT SELL
//     Price below market price: STOP_LOSS SELL, TAKE_PROFIT BUY
func (pp *PairProcessor) createOrder(
	orderType binance.OrderType, // MARKET, LIMIT, LIMIT_MAKER, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	sideType binance.SideType, // BUY, SELL
	timeInForce binance.TimeInForceType, // GTC, IOC, FOK
	quantity float64, // BTC for example if we buy or sell BTC
	quantityQty float64, // USDT for example if we buy or sell BTC
	// price for 1 BTC
	// it's price of order execution for LIMIT, LIMIT_MAKER
	// after execution of STOP_LOSS, TAKE_PROFIT, wil be created MARKET order
	// after execution of STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT wil be created LIMIT order with price of order execution from PRICE parameter
	price float64,
	// price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	stopPrice float64,
	// trailingDelta for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	// https://github.com/binance/binance-spot-api-docs/blob/master/faqs/trailing-stop-faq.md
	trailingDelta int,
	times int) (
	order *binance.CreateOrderResponse, err error) {
	if times == 0 {
		err = fmt.Errorf("can't create order")
		return
	}
	symbol, err := (*pp.pairInfo).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	if _, ok := pp.orderTypes[string(orderType)]; !ok && len(pp.orderTypes) != 0 {
		err = fmt.Errorf("order type %s is not supported for symbol %s", orderType, pp.pairInfo.Symbol)
		return
	}
	var (
		quantityRound = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))
		priceRound    = int(math.Log10(1 / utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))
	)
	service :=
		pp.client.NewCreateOrderService().
			NewOrderRespType(binance.NewOrderRespTypeRESULT).
			Symbol(string(binance.SymbolType(pp.pairInfo.Symbol))).
			Type(orderType).
			Side(sideType)
	// Additional mandatory parameters based on type:
	// Type	Additional mandatory parameters
	if orderType == binance.OrderTypeMarket {
		// MARKET	quantity or quoteOrderQty
		if quantity != 0 {
			service = service.
				Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
		} else if quantityQty != 0 {
			service = service.
				QuoteOrderQty(utils.ConvFloat64ToStr(quantityQty, quantityRound))
		} else {
			err = fmt.Errorf("quantity or quoteOrderQty must be set")
			return
		}
	} else if orderType == binance.OrderTypeLimit {
		// LIMIT	timeInForce, quantity, price
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == binance.OrderTypeLimitMaker {
		// LIMIT_MAKER	quantity, price
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
	} else if orderType == binance.OrderTypeStopLoss || orderType == binance.OrderTypeTakeProfit {
		// STOP_LOSS/TAKE_PROFIT quantity, stopPrice or trailingDelta
		service = service.
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(price, priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
		}
	} else if orderType == binance.OrderTypeStopLossLimit || orderType == binance.OrderTypeTakeProfitLimit {
		// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
		service = service.
			TimeInForce(timeInForce).
			Quantity(utils.ConvFloat64ToStr(quantity, quantityRound)).
			Price(utils.ConvFloat64ToStr(price, priceRound))
		if stopPrice != 0 {
			service = service.StopPrice(utils.ConvFloat64ToStr(price, priceRound))
		} else if trailingDelta != 0 {
			service = service.TrailingDelta(strconv.Itoa(trailingDelta))
		} else {
			err = fmt.Errorf("stopPrice or trailingDelta must be set")
			return
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
				if order.Symbol == pp.pairInfo.Symbol && order.Side == sideType && order.Price == utils.ConvFloat64ToStr(price, priceRound) {
					return &binance.CreateOrderResponse{
						Symbol:                   order.Symbol,
						OrderID:                  order.OrderID,
						ClientOrderID:            order.ClientOrderID,
						Price:                    order.Price,
						OrigQuantity:             order.OrigQuantity,
						ExecutedQuantity:         order.ExecutedQuantity,
						CummulativeQuoteQuantity: order.CummulativeQuoteQuantity,
						IsIsolated:               order.IsIsolated,
						Status:                   order.Status,
						TimeInForce:              order.TimeInForce,
						Type:                     order.Type,
						Side:                     order.Side,
					}, nil
				}
			}
		} else if apiError.Code == -1008 {
			time.Sleep(3 * time.Second)
			return pp.createOrder(orderType, sideType, timeInForce, quantity, quantityQty, price, stopPrice, trailingDelta, times-1)
		}
	}
	return
}

func (pp *PairProcessor) CreateOrder(
	orderType binance.OrderType, // MARKET, LIMIT, LIMIT_MAKER, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	sideType binance.SideType, // BUY, SELL
	timeInForce binance.TimeInForceType, // GTC, IOC, FOK
	quantity float64, // BTC for example if we buy or sell BTC
	quantityQty float64, // USDT for example if we buy or sell BTC
	// price for 1 BTC
	// it's price of order execution for LIMIT, LIMIT_MAKER
	// after execution of STOP_LOSS, TAKE_PROFIT, wil be created MARKET order
	// after execution of STOP_LOSS_LIMIT, TAKE_PROFIT_LIMIT wil be created LIMIT order with price of order execution from PRICE parameter
	price float64,
	// price for stop loss or take profit it's price of order execution for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	stopPrice float64,
	// trailingDelta for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT
	// https://github.com/binance/binance-spot-api-docs/blob/master/faqs/trailing-stop-faq.md
	trailingDelta int) (
	order *binance.CreateOrderResponse, err error) {
	return pp.createOrder(orderType, sideType, timeInForce, quantity, quantityQty, price, stopPrice, trailingDelta, 3)
}

func (pp *PairProcessor) UserDataEventStart(
	callBack func(event *binance.WsUserDataEvent),
	eventType ...binance.UserDataEventType) (resetEvent chan error, err error) {
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
	eventMap := make(map[binance.UserDataEventType]bool)
	for _, event := range eventType {
		eventMap[event] = true
	}
	wsHandler := func(event *binance.WsUserDataEvent) {
		if len(eventType) == 0 || eventMap[event.Event] {
			callBack(event)
		}
	}
	// Запускаємо стрім подій користувача
	var stopC chan struct{}
	_, stopC, err = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
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
				_, stopC, _ = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
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
					_, stopC, _ = binance.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
				}
			}
		}
	}()
	return
}

func (pp *PairProcessor) LimitUpdaterStream() {
	go func() {
		for {
			select {
			case <-time.After(pp.updateTime):
				pp.updateTime,
					pp.minuteOrderLimit,
					pp.dayOrderLimit,
					pp.minuteRawRequestLimit = LimitRead(pp.degree, []string{pp.pairInfo.Symbol}, pp.client)
			case <-pp.stop:
				close(pp.stop)
				return
			}
		}
	}()
}

func (pp *PairProcessor) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairProcessor) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func (pp *PairProcessor) CheckOrderType(orderType binance.OrderType) bool {
	_, ok := pp.orderTypes[string(orderType)]
	return ok
}

func (pp *PairProcessor) GetOpenOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOpenOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetAllOrders() (orders []*binance.Order, err error) {
	return pp.client.NewListOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetOrder(orderID int64) (order *binance.Order, err error) {
	return pp.client.NewGetOrderService().Symbol(pp.pairInfo.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelOrder(orderID int64) (order *binance.CancelOrderResponse, err error) {
	return pp.client.NewCancelOrderService().Symbol(pp.pairInfo.Symbol).OrderID(orderID).Do(context.Background())
}

func (pp *PairProcessor) CancelAllOrders() (orders *binance.CancelOpenOrdersResponse, err error) {
	return pp.client.NewCancelOpenOrdersService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
}

func (pp *PairProcessor) GetSymbol() *symbol_types.SpotSymbol {
	return pp.pairInfo
}

func (pp *PairProcessor) GetCurrentPrice() (float64, error) {
	price, err := pp.client.NewListPricesService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func (pp *PairProcessor) GetPair() string {
	return pp.pairInfo.Symbol
}

func (pp *PairProcessor) GetAccount() (account *binance.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetBaseAsset() (asset *binance.Balance, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Balances {
		if asset.Asset == pp.baseSymbol {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.baseSymbol)
}

func (pp *PairProcessor) GetTargetAsset() (asset *binance.Balance, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Balances {
		if asset.Asset == pp.targetSymbol {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.targetSymbol)
}

func (pp *PairProcessor) GetBaseBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
	return
}

func (pp *PairProcessor) GetTargetBalance() (balance float64, err error) {
	asset, err := pp.GetTargetAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
	return
}

func (pp *PairProcessor) GetFreeBalance() (balance float64) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return 0
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
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

func (pp *PairProcessor) GetLockedBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Locked) // Convert string to float64
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

func (pp *PairProcessor) Debug(fl, id string) {
	if logrus.GetLevel() == logrus.DebugLevel {
		orders, _ := pp.GetOpenOrders()
		logrus.Debugf("%s %s %s:", fl, id, pp.symbol.Symbol)
		for _, order := range orders {
			logrus.Debugf(" Open Order %v on price %v OrderSide %v Status %s", order.OrderID, order.Price, order.Side, order.Status)
		}
	}
}

func NewPairProcessor(
	client *binance.Client,
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
	debug bool,
	functions ...Functions) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = spot_exchange_info.Init(exchangeInfo, 3, client)
	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:       client,
		exchangeInfo: exchangeInfo,

		updateTime:            0,
		minuteOrderLimit:      &exchange_types.RateLimits{},
		dayOrderLimit:         &exchange_types.RateLimits{},
		minuteRawRequestLimit: &exchange_types.RateLimits{},

		stop: stop,

		pairInfo:     nil,
		orderTypes:   map[string]bool{},
		degree:       3,
		debug:        debug,
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}

	// Перевіряємо ліміти на ордери та запити
	pp.updateTime,
		pp.minuteOrderLimit,
		pp.dayOrderLimit,
		pp.minuteRawRequestLimit =
		LimitRead(degree, []string{symbol}, client)

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: symbol}).(*symbol_types.SpotSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	pp.orderTypes = make(map[string]bool, 0)
	for _, orderType := range pp.pairInfo.OrderTypes {
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

	if functions != nil {
		pp.testUp = functions[0].TestUp
		pp.testDown = functions[0].TestDown
		pp.nextPriceUp = functions[0].NextPriceUp
		pp.nextPriceDown = functions[0].NextPriceDown
		pp.nextQuantityUp = functions[0].NextQuantityUp
		pp.nextQuantityDown = functions[0].NextQuantityDown
	} else {
		pp.testUp = func(s, e float64) bool { return s < e }
		pp.testDown = func(s, e float64) bool { return s > e }
		pp.nextPriceUp = func(s float64, n int) float64 {
			return pp.roundPrice(s * math.Pow(1+deltaPrice, float64(2)))
		}
		pp.nextPriceDown = func(s float64, n int) float64 {
			return pp.roundPrice(s * math.Pow(1-deltaPrice, float64(2)))
		}
		pp.nextQuantityUp = func(s float64, n int) float64 {
			return pp.roundQuantity(s * (math.Pow(1+deltaQuantity, float64(2))))
		}
		pp.nextQuantityDown = func(s float64, n int) float64 {
			return pp.roundQuantity(s * (math.Pow(1+deltaQuantity, float64(2))))
		}
	}

	return
}
