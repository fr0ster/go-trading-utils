package futures_signals

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	deltaUp    = 0.0005
	deltaDown  = 0.0005
	degree     = 3
	limit      = 1000
	interval   = "1m"
	reloadTime = 500 * time.Millisecond
)

func printError() {
	if logrus.GetLevel() == logrus.DebugLevel {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			logrus.Errorf("Error occurred in file: %s at line: %d", file, line)
		} else {
			logrus.Errorf("Error occurred but could not get the caller information")
		}
	}
}

func RunFuturesHolding(wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return fmt.Errorf("it should be implemented for futures")
}

func RunScalpingHolding(
	client *futures.Client,
	pair string,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callbackRate float64,
	percentsToStopSettingNewOrder float64,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	return RunFuturesGridTrading(
		client,
		pair,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		percentsToStopSettingNewOrder,
		callbackRate,
		progression,
		quit,
		wg)
}

func getCallBackTrading(
	pairProcessor *PairProcessor,
	shortPositionNewOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	var (
		sideUp    futures.SideType
		typeUp    futures.OrderType
		priceUp   float64
		sideDown  futures.SideType
		typeDown  futures.OrderType
		priceDown float64
	)
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate {
			risk, _ := pairProcessor.GetPositionRisk()
			if (event.OrderTradeUpdate.Type == shortPositionNewOrderType ||
				event.OrderTradeUpdate.Type == longPositionNewOrderType) &&
				risk != nil &&
				event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
				logrus.Debugf("Futures %s: Order filled %v type %v on price %v/activation price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.Type,
					event.OrderTradeUpdate.LastFilledPrice,
					event.OrderTradeUpdate.ActivationPrice,
					event.OrderTradeUpdate.AccumulatedFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				if event.OrderTradeUpdate.Side == futures.SideTypeBuy {
					// Відкрили позицію long купівлею, закриваємо її продажем
					sideUp = futures.SideTypeBuy
					typeUp = longPositionDecOrderType
					priceUp = pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice))
					sideDown = futures.SideTypeSell
					typeDown = futures.OrderTypeStopMarket
					priceDown = pairProcessor.NextPriceDown(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice))
				} else if event.OrderTradeUpdate.Side == futures.SideTypeSell {
					// Відкрили позицію short продажею, закриваємо її купівлею
					sideUp = futures.SideTypeSell
					typeUp = shortPositionDecOrderType
					priceUp = pairProcessor.NextPriceUp(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice))
					sideDown = futures.SideTypeBuy
					typeDown = futures.OrderTypeTrailingStopMarket
					priceDown = pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice))
				}
				sellOrder, buyOrder, err := openPosition(
					sideUp,
					typeUp,
					sideDown,
					typeDown,
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					priceUp,
					priceUp,
					priceDown,
					priceDown,
					pairProcessor)
				if err != nil {
					logrus.Errorf("Can't open position: %v", err)
					printError()
					close(quit)
					return
				}
				logrus.Debugf("Futures %s: Open position order %v side %v type %v on price %v quantity %v status %v",
					pairProcessor.GetPair(),
					sellOrder.OrderID,
					sellOrder.Side,
					sellOrder.Type,
					sellOrder.Price,
					sellOrder.OrigQuantity,
					sellOrder.Status)
				logrus.Debugf("Futures %s: Open position order %v side %v type %v on price %v quantity %v status %v",
					pairProcessor.GetPair(),
					buyOrder.OrderID,
					buyOrder.Side,
					buyOrder.Type,
					buyOrder.Price,
					buyOrder.OrigQuantity,
					buyOrder.Status)
			} else if (event.OrderTradeUpdate.Type == shortPositionDecOrderType ||
				event.OrderTradeUpdate.Type == longPositionDecOrderType) &&
				event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
				// Створюємо початкові ордери на продаж та купівлю
				_, _, err := openPosition(
					futures.SideTypeSell,
					shortPositionNewOrderType,
					futures.SideTypeBuy,
					longPositionNewOrderType,
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
					pairProcessor)
				if err != nil {
					logrus.Errorf("Can't open position: %v", err)
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func RunFuturesTrading(
	client *futures.Client,
	symbol string,
	degree int,
	limit int,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callBackRate float64,
	shortPositionNewOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(
		client,
		symbol,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callBackRate,
		quit,
		progression)
	if err != nil {
		printError()
		return
	}

	price, err := pairProcessor.GetCurrentPrice()
	if err != nil {
		return err
	}
	_, quantityUp, _, _, quantityDown, _, err = pairProcessor.InitPositionGrid(10, price)
	if err != nil {
		err = fmt.Errorf("can't check position: %v", err)
		printError()
		return
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, quantityUp, quantityDown)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	_, err = pairProcessor.UserDataEventStart(
		getCallBackTrading(
			pairProcessor,                       // pairProcessor
			futures.OrderTypeLimit,              // shortPositionNewOrderType
			futures.OrderTypeTrailingStopMarket, // shortPositionDecOrderType
			futures.OrderTypeLimit,              // longPositionNewOrderType
			futures.OrderTypeTrailingStopMarket, // longPositionDecOrderType
			quit))                               // quit
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,
		shortPositionNewOrderType,
		futures.SideTypeBuy,
		shortPositionNewOrderType,
		quantityUp,
		quantityDown,
		initPriceUp,
		initPriceUp,
		initPriceDown,
		initPriceDown,
		pairProcessor)
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

// Створення ордера для розміщення в грід
func createOrder(
	pairProcessor *PairProcessor,
	side futures.SideType,
	orderType futures.OrderType,
	quantity,
	price float64,
	stopPrice float64,
	callbackRate float64,
	closePosition bool) (order *futures.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		orderType,                  // orderType
		side,                       // sideType
		futures.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		closePosition,              // closePosition
		price,                      // price
		stopPrice,                  // stopPrice
		callbackRate)               // callbackRate
	return
}

func IsOrdersOpened(grid *grid_types.Grid, pairProcessor *PairProcessor, pair *pairs_types.Pairs) (err error) {
	grid.Ascend(func(item btree.Item) bool {
		var orderOut *futures.Order
		record := item.(*grid_types.Record)
		if record.GetOrderId() != 0 {
			orderOut, err = pairProcessor.GetOrder(record.GetOrderId())
			if err != nil {
				return false
			}
			if orderOut == nil || orderOut.Status != futures.OrderStatusTypeNew {
				err = fmt.Errorf("futures %s: Order %v not found or status %v", pairProcessor.GetPair(), record.GetOrderId(), orderOut.Status)
			}
		}
		return true
	})
	return err
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	pairProcessor *PairProcessor,
	side futures.SideType,
	grid *grid_types.Grid,
	percentsToStopSettingNewOrder float64,
	order *grid_types.Record,
	quantity float64,
	locked float64,
	risk *futures.PositionRisk) (err error) {
	var (
		takerRecord *grid_types.Record
		takerOrder  *futures.CreateOrderResponse
	)
	delta_percent := func(currentPrice float64) float64 {
		return math.Abs((currentPrice - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
	}
	if side == futures.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			upPrice := pairProcessor.roundPrice(order.GetPrice() * (1 + pairProcessor.GetDeltaPrice()))
			if (pairProcessor.GetUpBound() == 0 || upPrice <= pairProcessor.GetUpBound()) &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pairProcessor.GetFreeBalance() &&
				locked <= pairProcessor.GetFreeBalance() {
				upOrder, err := createOrder(
					pairProcessor,
					futures.SideTypeSell,
					futures.OrderTypeLimit,
					quantity,
					upPrice,
					upPrice,
					0,
					false)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Sell order %v on price %v status %v quantity %v",
					pairProcessor.GetPair(), order.GetOrderId(), upOrder.OrderID, upPrice, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upRecord := grid_types.NewRecord(upOrder.OrderID, upPrice, quantity, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upRecord)
				order.SetUpPrice(upPrice) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status == futures.OrderStatusTypeFilled {
					takerOrder = upOrder
				}
			} else {
				if pairProcessor.GetUpBound() == 0 || upPrice > pairProcessor.GetUpBound() {
					logrus.Debugf("Futures %s: UpBound %v isn't 0 and price %v > UpBound %v",
						pairProcessor.GetPair(), pairProcessor.GetUpBound(), upPrice, pairProcessor.GetUpBound())
				} else if utils.ConvStrToFloat64(risk.IsolatedMargin) > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance())
				} else if locked > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pairProcessor.GetPair(), locked, pairProcessor.GetFreeBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель нижче
		downPrice, ok := grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		if ok && downPrice.GetOrderId() == 0 && downPrice.GetQuantity() <= 0 {
			// Створюємо ордер на купівлю
			downOrder, err := createOrder(
				pairProcessor,
				futures.SideTypeBuy,
				futures.OrderTypeLimit,
				quantity,
				order.GetDownPrice(),
				order.GetDownPrice(),
				0,
				false)
			if err != nil {
				printError()
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetQuantity(quantity)           // Записуємо кількість ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
				pairProcessor.GetPair(), order.GetOrderId(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status, quantity)
			if downOrder.Status == futures.OrderStatusTypeFilled {
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				pairProcessor,
				takerOrder.Side,
				grid,
				percentsToStopSettingNewOrder,
				order,
				quantity,
				locked,
				risk)
			if err != nil {
				printError()
				return err
			}
		}
	} else if side == futures.SideTypeBuy {
		// Якшо нижче немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetDownPrice() == 0 {
			// Створюємо ордер на купівлю
			downPrice := pairProcessor.roundPrice(order.GetPrice() * (1 - pairProcessor.GetDeltaPrice()))
			if (pairProcessor.GetLowBound() == 0 || downPrice >= pairProcessor.GetLowBound()) &&
				delta_percent(downPrice) >= percentsToStopSettingNewOrder &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pairProcessor.GetFreeBalance() &&
				locked <= pairProcessor.GetFreeBalance() {
				downOrder, err := createOrder(
					pairProcessor,
					futures.SideTypeBuy,
					futures.OrderTypeLimit,
					quantity,
					downPrice,
					downPrice,
					0,
					false)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
					pairProcessor.GetPair(), order.GetOrderId(), downOrder.OrderID, downPrice, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downRecord := grid_types.NewRecord(downOrder.OrderID, downPrice, quantity, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downRecord)
				order.SetDownPrice(downPrice) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status == futures.OrderStatusTypeFilled {
					takerRecord = downRecord
					takerOrder = downOrder
				}
			} else {
				if pairProcessor.GetLowBound() == 0 || downPrice < pairProcessor.GetLowBound() {
					logrus.Debugf("Futures %s: LowBound %v isn't 0 and price %v < LowBound %v",
						pairProcessor.GetPair(), pairProcessor.GetLowBound(), downPrice, pairProcessor.GetLowBound())
				} else if delta_percent(downPrice) < percentsToStopSettingNewOrder {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pairProcessor.GetPair(), risk.LiquidationPrice, delta_percent(downPrice), percentsToStopSettingNewOrder)
				} else if utils.ConvStrToFloat64(risk.IsolatedMargin) > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance())
				} else if locked > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pairProcessor.GetPair(), locked, pairProcessor.GetFreeBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель вище
		upRecord, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upRecord.GetOrderId() == 0 && upRecord.GetQuantity() <= 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrder(
				pairProcessor,
				futures.SideTypeSell,
				futures.OrderTypeLimit,
				quantity,
				order.GetUpPrice(),
				order.GetUpPrice(),
				0,
				false)
			if err != nil {
				printError()
				return err
			}
			if upOrder.Status == futures.OrderStatusTypeFilled {
				takerRecord = upRecord
				takerOrder = upOrder
			}
			upRecord.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upRecord.SetQuantity(quantity)            // Записуємо кількість ордера в грід
			upRecord.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: From order %v Set Sell order %v on price %v status %v quantity %v",
				pairProcessor.GetPair(), order.GetOrderId(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				pairProcessor,
				takerOrder.Side,
				grid,
				percentsToStopSettingNewOrder,
				takerRecord,
				quantity,
				locked,
				risk)
			if err != nil {
				printError()
				return err
			}
		}
	}
	return
}

func initVars(
	pairProcessor *PairProcessor) (
	price float64,
	priceUp,
	priceDown float64,
	minNotional float64,
	err error) {
	symbol, err := pairProcessor.GetFuturesSymbol()
	if err != nil {
		return
	}
	// Отримання середньої ціни
	price, _ = pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
	price = pairProcessor.roundPrice(price)
	priceUp = pairProcessor.NextPriceUp(price)
	priceDown = pairProcessor.NextPriceDown(price)
	minNotional = utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
	return
}

func openPosition(
	sideUp futures.SideType,
	orderTypeUp futures.OrderType,
	sideDown futures.SideType,
	orderTypeDown futures.OrderType,
	quantityUp float64,
	quantityDown float64,
	priceUp float64,
	stopPriceUp float64,
	priceDown float64,
	stopPriceDown float64,
	pairProcessor *PairProcessor) (upOrder, downOrder *futures.CreateOrderResponse, err error) {
	err = pairProcessor.CancelAllOrders()
	if err != nil {
		printError()
		return
	}
	// Створюємо ордери на продаж
	upOrder, err = createOrder(
		pairProcessor,
		sideUp,
		orderTypeUp,
		quantityUp,
		priceUp,
		stopPriceUp,
		pairProcessor.GetCallbackRate(),
		false)
	if err != nil {
		logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
			pairProcessor.GetPair(), sideUp, orderTypeUp, priceUp, quantityUp, pairProcessor.GetCallbackRate())
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
		pairProcessor.GetPair(), sideUp, orderTypeUp, priceUp, quantityUp, pairProcessor.GetCallbackRate(), upOrder.Status)
	// Створюємо ордери на купівлю
	downOrder, err = createOrder(
		pairProcessor,
		sideDown,
		orderTypeDown,
		quantityDown,
		priceDown,
		stopPriceDown,
		pairProcessor.GetCallbackRate(),
		false)
	if err != nil {
		logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
			pairProcessor.GetPair(), sideDown, orderTypeDown, priceDown, quantityDown, pairProcessor.GetCallbackRate())
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
		pairProcessor.GetPair(), sideDown, orderTypeDown, priceDown, quantityDown, pairProcessor.GetCallbackRate(), downOrder.Status)
	return
}

func marginBalancing(
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor) (err error) {
	// Балансування маржі як треба
	if utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		delta := pairProcessor.roundPrice(pairProcessor.GetFreeBalance()) - pairProcessor.roundPrice(utils.ConvStrToFloat64(risk.IsolatedMargin))
		if delta != 0 {
			if delta > 0 && delta < pairProcessor.GetFreeBalance() {
				err = pairProcessor.SetPositionMargin(delta, 1)
				logrus.Debugf("Futures %s: IsolatedMargin %v < current position balance %v and we have enough free %v",
					pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance(), pairProcessor.GetFreeBalance())
			}
		}
	}
	return
}

func initGrid(
	pairProcessor *PairProcessor,
	price float64,
	quantity float64,
	sellOrder,
	buyOrder *futures.CreateOrderResponse) (grid *grid_types.Grid, err error) {
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pairProcessor.GetPair())
	grid = grid_types.New()
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, 0, pairProcessor.roundPrice(price*(1+pairProcessor.GetDeltaPrice())), pairProcessor.roundPrice(price*(1-pairProcessor.GetDeltaPrice())), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pairProcessor.GetPair(), price)
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, pairProcessor.roundPrice(price*(1+pairProcessor.GetDeltaPrice())), quantity, 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pairProcessor.GetPair(), pairProcessor.roundPrice(price*(1+pairProcessor.GetDeltaPrice())))
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, pairProcessor.roundPrice(price*(1-pairProcessor.GetDeltaPrice())), quantity, price, 0, types.SideTypeBuy))
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	return
}

func getCallBack_v1(
	// config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	percentsToStopSettingNewOrder float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		currentPrice float64
		risk         *futures.PositionRisk
		err          error
	)
	return func(event *futures.WsUserDataEvent) {
		if grid == nil {
			return
		}
		if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				grid.Lock()
				logrus.Debugf("Futures %s: Order %v on price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
				// Знаходимо у гріді на якому був виконаний ордер
				order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
				if ok {
					orderId := order.GetOrderId()
					locked, _ = pairProcessor.GetLockedBalance()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						grid.Unlock()
						printError()
						close(quit)
						return
					}
					// Балансування маржі як треба
					err = marginBalancing(risk, pairProcessor)
					if err != nil {
						grid.Unlock()
					}
					err = processOrder(
						pairProcessor,
						event.OrderTradeUpdate.Side,
						grid,
						percentsToStopSettingNewOrder,
						order,
						quantity,
						locked,
						risk)
					if err != nil {
						grid.Unlock()
						pairProcessor.CancelAllOrders()
						printError()
						close(quit)
						return
					}
					grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pairProcessor.GetPair())
					grid.Unlock()
				}
			}
		}
	}
}

func RunFuturesGridTrading(
	client *futures.Client,
	pair string,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	percentsToStopSettingNewOrder float64,
	callbackRate float64,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		initPrice     float64
		initPriceUp   float64
		initPriceDown float64
		quantity      float64
		quantityUp    float64
		quantityDown  float64
		minNotional   float64
		grid          *grid_types.Grid
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо обробник пари
	pairProcessor, err := NewPairProcessor(
		client,
		pair,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callbackRate,
		quit,
		progression)
	if err != nil {
		printError()
		return
	}
	initPrice, initPriceUp, initPriceDown, minNotional, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pairProcessor.GetFreeBalance(), pairProcessor.GetLimitOnTransaction())
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v1(
			pairProcessor,
			grid,
			percentsToStopSettingNewOrder,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(
		futures.SideTypeSell,   // sideUp
		futures.OrderTypeLimit, // typeUp
		futures.SideTypeBuy,    // sideDown
		futures.OrderTypeLimit, // typeDown
		quantityUp,             // quantityUp
		quantityDown,           // quantityDown
		initPriceUp,            // priceUp
		initPriceUp,            // stopPriceUp
		initPriceDown,          // priceDown
		initPriceDown,          // stopPriceDown
		pairProcessor)          // pairProcessor
	if err != nil {
		printError()
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pairProcessor, initPrice, quantity, sellOrder, buyOrder)
	if err != nil {
		printError()
		return err
	}
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

func getCallBack_v2(
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	percentsToStopSettingNewOrder float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		currentPrice float64
		risk         *futures.PositionRisk
		err          error
	)
	return func(event *futures.WsUserDataEvent) {
		if grid == nil {
			return
		}
		if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			grid.Lock()
			// Знаходимо у гріді на якому був виконаний ордер
			currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
			order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
			if !ok {
				printError()
				logrus.Errorf("we didn't work with order on price level %v before: %v", currentPrice, event.OrderTradeUpdate.ID)
				return
			}
			orderId := order.GetOrderId()
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				logrus.Debugf("Futures %s: Order %v on price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				locked, _ = pairProcessor.GetLockedBalance()
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				// Балансування маржі як треба
				err = marginBalancing(risk, pairProcessor)
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				err = processOrder(
					pairProcessor,
					event.OrderTradeUpdate.Side,
					grid,
					percentsToStopSettingNewOrder,
					order,
					quantity,
					locked,
					risk)
				if err != nil {
					grid.Unlock()
					pairProcessor.CancelAllOrders()
					printError()
					close(quit)
					return
				}
				grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pairProcessor.GetPair())
			}
			grid.Unlock()
		}
	}
}

func RunFuturesGridTradingV2(
	// config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callbackRate float64,
	percentsToStopSettingNewOrder float64,
	quit chan struct{},
	progression pairs_types.ProgressionType,
	wg *sync.WaitGroup) (err error) {
	var (
		initPrice     float64
		initPriceUp   float64
		initPriceDown float64
		quantity      float64
		quantityUp    float64
		quantityDown  float64
		minNotional   float64
		grid          *grid_types.Grid
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(
		client,
		pair.GetPair(),
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callbackRate,
		quit,
		progression)
	if err != nil {
		printError()
		return
	}
	initPrice, initPriceUp, initPriceDown, minNotional, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current limitOnTransaction %v",
			minNotional, pairProcessor.GetLimitOnTransaction())
	}
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v2(
			pairProcessor,
			grid,
			percentsToStopSettingNewOrder,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(
		futures.SideTypeSell,   // sideUp
		futures.OrderTypeLimit, // typeUp
		futures.SideTypeBuy,    // sideDown
		futures.OrderTypeLimit, // typeDown
		quantityUp,             // quantityUp
		quantityDown,           // quantityDown
		initPriceUp,            // priceUp
		initPriceUp,            // stopPriceUp
		initPriceDown,          // priceDown
		initPriceDown,          // stopPriceDown
		pairProcessor)          // pairProcessor
	if err != nil {
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pairProcessor, initPrice, quantity, sellOrder, buyOrder)
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

func getCallBack_v3(
	pairProcessor *PairProcessor,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			// Знаходимо у гріді на якому був виконаний ордер
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				logrus.Debugf("Futures %s: Order filled %v type %v on price %v/activation price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.Type,
					event.OrderTradeUpdate.LastFilledPrice,
					event.OrderTradeUpdate.ActivationPrice,
					event.OrderTradeUpdate.AccumulatedFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				risk, err := pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					pairProcessor.CancelAllOrders()
					close(quit)
					return
				}
				// Балансування маржі як треба
				marginBalancing(risk, pairProcessor)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v3(
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					event.OrderTradeUpdate.Side,
					shortPositionNewOrderType,
					shortPositionIncOrderType,
					shortPositionDecOrderType,
					longPositionNewOrderType,
					longPositionIncOrderType,
					longPositionDecOrderType,
					pairProcessor)
				if err != nil {
					logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func createNextPair_v3(
	LastExecutedPrice float64,
	AccumulatedFilledQty float64,
	LastExecutedSide futures.SideType,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	pairProcessor *PairProcessor) (err error) {
	var (
		risk         *futures.PositionRisk
		upPrice      float64
		downPrice    float64
		upQuantity   float64
		downQuantity float64
		sellOrder    *futures.CreateOrderResponse
		buyOrder     *futures.CreateOrderResponse
		callBackRate float64 = pairProcessor.GetCallbackRate()
	)
	risk, _ = pairProcessor.GetPositionRisk()
	free := pairProcessor.GetFreeBalance()
	currentPrice, _ := pairProcessor.GetCurrentPrice()
	positionVal := utils.ConvStrToFloat64(risk.PositionAmt) * currentPrice / float64(pairProcessor.GetLeverage())
	if AccumulatedFilledQty == utils.ConvStrToFloat64(risk.PositionAmt) {
		_, upQuantity, _, _, downQuantity, _, err = pairProcessor.InitPositionGrid(10, LastExecutedPrice)
		if err != nil {
			err = fmt.Errorf("can't check position: %v", err)
			printError()
			return
		}
	}
	if positionVal < 0 { // Маємо позицію short
		if positionVal >= -free {
			// Виконаний ордер був на продаж, тобто збільшив позицію short
			if LastExecutedSide == futures.SideTypeSell {
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				upPrice, upQuantity, err = pairProcessor.NextUp(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position: %v", err)
					return
				}
				// Створюємо ордер на продаж, тобто збільшуємо позицію short
				// Створюємо ордер на купівлю, тобто скорочуємо позицію short
				sellOrder, buyOrder, err = openPosition(
					futures.SideTypeSell,      // sideUp
					shortPositionIncOrderType, // typeUp
					futures.SideTypeBuy,       // sideDown
					shortPositionDecOrderType, // typeDown
					upQuantity,                // quantityUp
					AccumulatedFilledQty,      // quantityDown
					upPrice,                   // priceUp
					upPrice,                   // stopPriceUp
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // priceDown
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // stopPriceDown
					pairProcessor) // pairProcessor
				// Виконаний ордер був на купівлю, тобто скоротив позицію short
				// Обробляємо розворот курсу
			} else if LastExecutedSide == futures.SideTypeBuy {
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				upPrice, upQuantity, err = pairProcessor.NextDown(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position: %v", err)
					return
				}
				// Створюємо ордер на продаж, тобто збільшуємо позицію short
				// Створюємо ордер на купівлю, тобто скорочуємо позицію short
				sellOrder, buyOrder, err = openPosition(
					futures.SideTypeSell,      // sideUp
					shortPositionIncOrderType, // typeUp
					futures.SideTypeBuy,       // sideDown
					shortPositionDecOrderType, // typeDown
					upQuantity,                // quantityUp
					AccumulatedFilledQty,      // quantityDown
					upPrice,                   // priceUp
					upPrice,                   // stopPriceUp
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // priceDown
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // stopPriceDown
					pairProcessor) // pairProcessor

			}
			if err != nil {
				logrus.Errorf("Can't open position: %v", err)
				printError()
				return
			}
			logrus.Debugf("Futures %s: Sell type %v Quantity Up %v * upPrice %v = %v, status %v",
				pairProcessor.GetPair(),
				longPositionIncOrderType,
				upQuantity,
				upPrice,
				upQuantity*upPrice,
				sellOrder.Status)
			logrus.Debugf("Futures %s: Buy type %v Quantity Down %v * downPrice %v = %v, status %v",
				pairProcessor.GetPair(),
				longPositionDecOrderType,
				AccumulatedFilledQty,
				pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				AccumulatedFilledQty*pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				buyOrder.Status)
		} else {
			// Створюємо ордер на купівлю, тобто скорочуємо позицію short
			buyOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeBuy,
				shortPositionDecOrderType,
				AccumulatedFilledQty,
				pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				callBackRate,
				false)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, stopPrice %v, callbackRate %v, status %v",
					pairProcessor.GetPair(),
					futures.SideTypeSell,
					shortPositionDecOrderType,
					AccumulatedFilledQty,
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
					pairProcessor.NextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
					callBackRate,
					buyOrder.Status)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		}
	} else if positionVal > 0 { // Маємо позицію long
		if positionVal <= free {
			// Виконаний ордер був на купівлю, тобто збільшив позицію long
			if LastExecutedSide == futures.SideTypeBuy {
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				downPrice, downQuantity, err = pairProcessor.NextDown(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position: %v", err)
					return
				}
				// Створюємо ордер на продаж, тобто скорочуємо позицію long
				// Створюємо ордер на купівлю, тобто збільшуємо позицію long
				sellOrder, buyOrder, err = openPosition(
					futures.SideTypeSell,     // sideUp
					longPositionIncOrderType, // typeUp
					futures.SideTypeBuy,      // sideDown
					longPositionDecOrderType, // typeDown
					AccumulatedFilledQty,     // quantityUp
					downQuantity,             // quantityDown
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // priceUp
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // stopPriceUp
					downPrice,     // priceDown
					downPrice,     // stopPriceDown
					pairProcessor) // pairProcessor
				// Виконаний ордер був на продаж, тобто скоротив позицію long
				// Обробляємо розворот курсу
			} else if LastExecutedSide == futures.SideTypeSell {
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				downPrice, downQuantity, err = pairProcessor.NextUp(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position: %v", err)
					return
				}
				// Створюємо ордер на продаж, тобто скорочуємо позицію long
				// Створюємо ордер на купівлю, тобто збільшуємо позицію long
				sellOrder, buyOrder, err = openPosition(
					futures.SideTypeSell,     // sideUp
					longPositionIncOrderType, // typeUp
					futures.SideTypeBuy,      // sideDown
					longPositionDecOrderType, // typeDown
					AccumulatedFilledQty,     // quantityUp
					downQuantity,             // quantityDown
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // priceUp
					pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)), // stopPriceUp
					downPrice,     // priceDown
					downPrice,     // stopPriceDown
					pairProcessor) // pairProcessor
			}
			if err != nil {
				logrus.Errorf("Can't open position: %v", err)
				printError()
				return
			}
			logrus.Debugf("Futures %s: Sell type %v Quantity Up %v * upPrice %v = %v, status %v",
				pairProcessor.GetPair(),
				longPositionIncOrderType,
				AccumulatedFilledQty,
				pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				AccumulatedFilledQty*pairProcessor.NextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice)),
				sellOrder.Status)
			logrus.Debugf("Futures %s: Buy type %v Quantity Down %v * downPrice %v = %v, status %v",
				pairProcessor.GetPair(),
				longPositionDecOrderType,
				downQuantity,
				downPrice,
				downPrice*downQuantity,
				buyOrder.Status)
		} else {
			// Створюємо ордер на продаж, тобто скорочуємо позицію long
			sellOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeSell,
				longPositionDecOrderType,
				upQuantity,
				upPrice,
				upPrice,
				callBackRate,
				false)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, callbackRate %v, status %v",
					pairProcessor.GetPair(),
					futures.SideTypeSell,
					longPositionDecOrderType,
					upQuantity,
					upPrice,
					callBackRate,
					sellOrder.Status)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		}
	} else { // Немає позиції, відкриваємо нову
		// Відкриваємо нову позицію
		// Визначаємо ціну для нових ордерів
		// Визначаємо кількість для нових ордерів
		openPosition(
			futures.SideTypeBuy,       // sideUp
			shortPositionNewOrderType, // typeUp
			futures.SideTypeSell,      // sideDown
			longPositionNewOrderType,  // typeDown
			upQuantity,                // quantityUp
			downQuantity,              // quantityDown
			upPrice,                   // priceUp
			upPrice,                   // stopPriceUp
			downPrice,                 // priceDown
			downPrice,                 // stopPriceDown
			pairProcessor)             // pairProcessor
		logrus.Debugf("Futures %s: Buy type %v Quantity Up %v * upPrice %v = %v",
			pairProcessor.GetPair(), shortPositionNewOrderType, upQuantity, upPrice, upQuantity*upPrice)
		logrus.Debugf("Futures %s: Sell type %v Quantity Down %v * downPrice %v = %v",
			pairProcessor.GetPair(), longPositionNewOrderType, downQuantity, downPrice, downQuantity*downPrice)
	}
	return
}

// Працюємо лімітними ордерами (але можливо зменьшувати позицію будемо і TakeProfit ордером),
// відкриваємо ордера на продаж та купівлю з однаковою кількістью
// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
// після 5 наприклад ордера ставимо на більшу відстань
func RunFuturesGridTradingV3(
	client *futures.Client,
	pair string,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callbackRate float64,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		// minNotional   float64
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true

	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(
		client,
		pair,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callbackRate,
		quit,
		progression)
	if err != nil {
		printError()
		return
	}
	price, err := pairProcessor.GetCurrentPrice()
	if err != nil {
		return err
	}
	_, quantityUp, _, _, quantityDown, _, err = pairProcessor.InitPositionGrid(10, price)
	if err != nil {
		err = fmt.Errorf("can't check position: %v", err)
		printError()
		return
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, quantityUp, quantityDown)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v3(
			pairProcessor,
			shortPositionNewOrderType,
			shortPositionIncOrderType,
			shortPositionDecOrderType,
			longPositionNewOrderType,
			longPositionIncOrderType,
			longPositionDecOrderType,
			maintainedOrders,
			quit))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,      // sideUp
		shortPositionNewOrderType, // typeUp
		futures.SideTypeBuy,       // sideDown
		longPositionNewOrderType,  // typeDown
		quantityUp,                // quantityUp
		quantityDown,              // quantityDown
		initPriceUp,               // priceUp
		initPriceUp,               // stopPriceUp
		initPriceDown,             // priceDown
		initPriceDown,             // stopPriceDown
		pairProcessor)             // pairProcessor
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

func getCallBack_v4(
	pairProcessor *PairProcessor,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			// Знаходимо у гріді на якому був виконаний ордер
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				logrus.Debugf("Futures %s: Order filled %v type %v on price %v/activation price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.Type,
					event.OrderTradeUpdate.LastFilledPrice,
					event.OrderTradeUpdate.ActivationPrice,
					event.OrderTradeUpdate.AccumulatedFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				risk, err := pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					pairProcessor.CancelAllOrders()
					close(quit)
					return
				}
				logrus.Debugf("Futures %s: Event OrderTradeUpdate: Side %v Type %v OriginalPrice %v, OriginalQty %v, LastFilledPrice %v, LastFilledQty %v",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Type,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.OriginalQty,
					event.OrderTradeUpdate.LastFilledPrice,
					event.OrderTradeUpdate.LastFilledQty)
				// Балансування маржі як треба
				marginBalancing(risk, pairProcessor)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v4(
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),      // LastExecutedPrice
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty), // AccumulatedFilledQty
					shortPositionNewOrderType, // shortPositionNewOrderType
					shortPositionIncOrderType, // shortPositionIncOrderType
					shortPositionDecOrderType, // shortPositionDecOrderType
					longPositionNewOrderType,  // longPositionNewOrderType
					longPositionIncOrderType,  // longPositionIncOrderType
					longPositionDecOrderType,  // longPositionDecOrderType
					pairProcessor)             // pairProcessor
				if err != nil {
					logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func createNextPair_v4(
	LastExecutedPrice float64,
	AccumulatedFilledQty float64,
	// LastExecutedSide futures.SideType,
	// quantity float64,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	pairProcessor *PairProcessor) (err error) {
	var (
		risk             *futures.PositionRisk
		upPrice          float64
		downPrice        float64
		upQuantity       float64
		downQuantity     float64
		callBackRate     float64 = pairProcessor.GetCallbackRate()
		createdOrderUp   bool    = false
		createdOrderDown bool    = false
		sellOrder        *futures.CreateOrderResponse
		buyOrder         *futures.CreateOrderResponse
	)
	risk, _ = pairProcessor.GetPositionRisk()
	if AccumulatedFilledQty == utils.ConvStrToFloat64(risk.PositionAmt) {
		_, upQuantity, _, _, downQuantity, _, err = pairProcessor.InitPositionGrid(10, LastExecutedPrice)
		if err != nil {
			err = fmt.Errorf("can't check position: %v", err)
			printError()
			return
		}
	}
	getClosePosition := func(risk *futures.PositionRisk) (up, down bool) {
		// Визначаємо кількість для нових ордерів коли позиція від'ємна
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			up = false
			down = true
			// Визначаємо кількість для нових ордерів коли позиція позитивна
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			up = true
			down = false
			// Визначаємо кількість для нових ордерів коли позиція нульова
		} else {
			up = false
			down = false
		}
		return
	}
	upClosePosition, downClosePosition := getClosePosition(risk)
	if pairProcessor.GetUpBound() != 0 && upPrice <= pairProcessor.GetUpBound() && upQuantity > 0 {
		if upClosePosition {
			sellOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeSell,
				futures.OrderTypeTrailingStopMarket,
				upQuantity,
				upPrice,
				upPrice,
				callBackRate,
				upClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, callbackRate %v",
					pairProcessor.GetPair(), futures.SideTypeSell, futures.OrderTypeTrailingStopMarket, upQuantity, upPrice, callBackRate)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		} else {
			sellOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeSell,
				futures.OrderTypeLimit,
				upQuantity,
				upPrice,
				upPrice,
				0,
				upClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, callbackRate %v",
					pairProcessor.GetPair(), futures.SideTypeSell, futures.OrderTypeTrailingStopMarket, upQuantity, upPrice, callBackRate)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		}
		logrus.Debugf("Futures %s: Create Sell order type %v on price %v quantity %v status %v",
			pairProcessor.GetPair(), sellOrder.Type, upPrice, upQuantity, sellOrder.Status)
		if sellOrder.Status == futures.OrderStatusTypeFilled {
			pairProcessor.CancelAllOrders()
			return createNextPair_v4(
				upPrice,                   // LastExecutedPrice
				upQuantity,                // AccumulatedFilledQty
				shortPositionNewOrderType, // shortPositionNewOrderType
				shortPositionIncOrderType, // shortPositionIncOrderType
				shortPositionDecOrderType, // shortPositionDecOrderType
				longPositionNewOrderType,  // longPositionNewOrderType
				longPositionIncOrderType,  // longPositionIncOrderType
				longPositionDecOrderType,  // longPositionDecOrderType
				pairProcessor)             // pairProcessor
		}
		createdOrderUp = true
	} else {
		if upQuantity <= 0 {
			logrus.Debugf("Futures %s: upQuantity %v less than 0", pairProcessor.GetPair(), upQuantity)
		} else {
			logrus.Debugf("Futures %s: upPrice %v more than upBound %v",
				pairProcessor.GetPair(), upPrice, pairProcessor.GetUpBound())
		}
	}
	// Створюємо ордер на купівлю
	if pairProcessor.GetLowBound() != 0 && downPrice >= pairProcessor.GetLowBound() && downQuantity > 0 {
		if downClosePosition {
			buyOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeBuy,
				futures.OrderTypeTrailingStopMarket,
				downQuantity,
				downPrice,
				downPrice,
				callBackRate,
				downClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Buy order: side %v, type %v, quantity %v, price %v, callbackRate %v",
					pairProcessor.GetPair(), futures.SideTypeBuy, futures.OrderTypeTrailingStopMarket, upQuantity, upPrice, callBackRate)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		} else {
			buyOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeBuy,
				futures.OrderTypeLimit,
				downQuantity,
				downPrice,
				downPrice,
				0,
				downClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Buy order: side %v, type %v, quantity %v, price %v, callbackRate %v",
					pairProcessor.GetPair(), futures.SideTypeBuy, futures.OrderTypeTrailingStopMarket, upQuantity, upPrice, callBackRate)
				logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
				printError()
				return
			}
		}
		logrus.Debugf("Futures %s: Create Buy order type %v on price %v quantity %v status %v",
			pairProcessor.GetPair(), buyOrder.Type, downPrice, downQuantity, buyOrder.Status)
		if buyOrder.Status == futures.OrderStatusTypeFilled {
			pairProcessor.CancelAllOrders()
			return createNextPair_v4(
				upPrice,                   // LastExecutedPrice
				upQuantity,                // AccumulatedFilledQty
				shortPositionNewOrderType, // shortPositionNewOrderType
				shortPositionIncOrderType, // shortPositionIncOrderType
				shortPositionDecOrderType, // shortPositionDecOrderType
				longPositionNewOrderType,  // longPositionNewOrderType
				longPositionIncOrderType,  // longPositionIncOrderType
				longPositionDecOrderType,  // longPositionDecOrderType
				pairProcessor)             // pairProcessor
		}
		createdOrderDown = true
		logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v", pairProcessor.GetPair(), downPrice, downQuantity)
	} else {
		if downQuantity <= 0 {
			logrus.Debugf("Futures %s: downQuantity %v less than 0", pairProcessor.GetPair(), downQuantity)
		} else {
			logrus.Debugf("Futures %s: downPrice %v less than lowBound %v",
				pairProcessor.GetPair(), downPrice, pairProcessor.GetLowBound())
		}
	}
	if !createdOrderUp && !createdOrderDown {
		logrus.Debugf("Futures %s: Orders was not created", pairProcessor.GetPair())
		printError()
		return fmt.Errorf("orders were not created")
	}
	return
}

// Працюємо лімітними та TakeProfit/TrailingStop ордерами,
// відкриваємо лімітний ордер на збільшення, а закриваємо всю позицію TakeProfit/TrailingStop або лімітним ордером
// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
// після 5 наприклад ордера ставимо на більшу відстань
// або статично відкриємо ордери на продаж та купівлю з однаковою кількістью та с однаковим шагом
func RunFuturesGridTradingV4(
	client *futures.Client,
	pair string,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callbackRate float64,
	shortPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionNewOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		// minNotional   float64
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(
		client,
		pair,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callbackRate,
		quit,
		progression)
	if err != nil {
		printError()
		return
	}
	price, err := pairProcessor.GetCurrentPrice()
	if err != nil {
		return err
	}
	_, quantityUp, _, _, quantityDown, _, err = pairProcessor.InitPositionGrid(10, price)
	if err != nil {
		err = fmt.Errorf("can't check position: %v", err)
		printError()
		return
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, quantityUp, quantityDown)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v4(
			pairProcessor,
			shortPositionNewOrderType,
			shortPositionIncOrderType,
			shortPositionDecOrderType,
			longPositionNewOrderType,
			longPositionIncOrderType,
			longPositionDecOrderType,
			maintainedOrders,
			quit))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,      // sideUp
		shortPositionNewOrderType, // typeUp
		futures.SideTypeBuy,       // sideDown
		longPositionNewOrderType,  // typeDown
		quantityUp,                // quantityUp
		quantityDown,              // quantityDown
		initPriceUp,               // priceUp
		initPriceUp,               // stopPriceUp
		initPriceDown,             // priceDown
		initPriceDown,             // stopPriceDown
		pairProcessor)             // pairProcessor
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	quit chan struct{},
	debug bool,
	wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		var err error
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			err = fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			err = RunFuturesHolding(wg)

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			err = RunScalpingHolding(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetProgression(), // progression
				quit,                  // quit
				wg)                    // wg

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			err = RunFuturesTrading(
				client,                              // client
				pair.GetPair(),                      // pair
				degree,                              // degree
				limit,                               // limit
				pair.GetLimitOnPosition(),           // limitOnPosition
				pair.GetLimitOnTransaction(),        // limitOnTransaction
				pair.GetUpBound(),                   // upBound
				pair.GetLowBound(),                  // lowBound
				pair.GetDeltaPrice(),                // deltaPrice
				pair.GetDeltaQuantity(),             // deltaQuantity
				pair.GetMarginType(),                // marginType
				pair.GetLeverage(),                  // leverage
				pair.GetCallbackRate(),              // callbackRate
				futures.OrderTypeLimit,              // shortPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // shortPositionDecOrderType
				futures.OrderTypeLimit,              // longPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // longPositionDecOrderType
				pair.GetProgression(),               // progression
				quit,                                // quit
				wg)

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			err = RunFuturesGridTrading(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetProgression(), // progression
				quit,                  // quit
				wg)                    // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV2 {
			err = RunFuturesGridTradingV2(
				client,                       // client
				pair,                         // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				quit,                  // quit
				pair.GetProgression(), // progression
				wg)                    // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV3 {
			err = RunFuturesGridTradingV3(
				client,                              // client
				pair.GetPair(),                      // pair
				pair.GetLimitOnPosition(),           // limitOnPosition
				pair.GetLimitOnTransaction(),        // limitOnTransaction
				pair.GetUpBound(),                   // upBound
				pair.GetLowBound(),                  // lowBound
				pair.GetDeltaPrice(),                // deltaPrice
				pair.GetDeltaQuantity(),             // deltaQuantity
				pair.GetMarginType(),                // marginType
				pair.GetLeverage(),                  // leverage
				pair.GetCallbackRate(),              // callbackRate
				futures.OrderTypeLimit,              // shortPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // shortPositionIncOrderType
				futures.OrderTypeTrailingStopMarket, // shortPositionDecOrderType
				futures.OrderTypeLimit,              // longPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // longPositionIncOrderType
				futures.OrderTypeTrailingStopMarket, // longPositionDecOrderType
				pair.GetProgression(),               // progression
				quit,                                // quit
				wg)                                  // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV4 {
			err = RunFuturesGridTradingV4(
				client,                              // client
				pair.GetPair(),                      // pair
				pair.GetLimitOnPosition(),           // limitOnPosition
				pair.GetLimitOnTransaction(),        // limitOnTransaction
				pair.GetUpBound(),                   // upBound
				pair.GetLowBound(),                  // lowBound
				pair.GetDeltaPrice(),                // deltaPrice
				pair.GetDeltaQuantity(),             // deltaQuantity
				pair.GetMarginType(),                // marginType
				pair.GetLeverage(),                  // leverage
				pair.GetCallbackRate(),              // callbackRate
				futures.OrderTypeLimit,              // shortPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // shortPositionIncOrderType
				futures.OrderTypeTrailingStopMarket, // shortPositionDecOrderType
				futures.OrderTypeLimit,              // longPositionNewOrderType
				futures.OrderTypeTrailingStopMarket, // longPositionIncOrderType
				futures.OrderTypeTrailingStopMarket, // longPositionDecOrderType
				pair.GetProgression(),               // progression
				quit,                                // quit
				wg)                                  // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV5 {
			err = RunFuturesGridTradingV3(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetCallbackRate(),       // callbackRate
				futures.OrderTypeLimit,       // shortPositionNewOrderType
				futures.OrderTypeTakeProfit,  // shortPositionIncOrderType
				futures.OrderTypeTakeProfit,  // shortPositionDecOrderType
				futures.OrderTypeLimit,       // longPositionNewOrderType
				futures.OrderTypeTakeProfit,  // longPositionIncOrderType
				futures.OrderTypeTakeProfit,  // longPositionDecOrderType
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg)                           // wg

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			err = fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
		if err != nil {
			logrus.Error(err)
			close(quit)
		}
	}()
}
