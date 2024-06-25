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
	config *config_types.ConfigFile,
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
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	return RunFuturesGridTrading(
		config,
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
		callbackRate,
		quit,
		wg)
}

func getCallBackTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pairProcessor *PairProcessor,
	tickSizeExp int,
	minNotional float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			logrus.Debugf("Futures %s: Order %v filled", pairProcessor.GetPair(), event.OrderTradeUpdate.ID)
		}
	}
}

func RunFuturesTrading(
	config *config_types.ConfigFile,
	symbol string,
	client *futures.Client,
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
	quit chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	var (
		initPrice     float64
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		minNotional   float64
		tickSizeExp   int
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо стрім подій
	pairProcessor, err = initRun(
		symbol,
		client,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		callBackRate,
		quit)
	if err != nil {
		return err
	}
	_, initPrice, _, _, minNotional, tickSizeExp, _, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current limitOnTransaction %v",
			minNotional, pairProcessor.GetLimitOnTransaction())
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.InitPositionGrid(10, initPrice)
	if err != nil {
		logrus.Errorf("Can't check position: %v", err)
		close(quit)
		return
	}
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		printError()
		close(quit)
		return err
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		initPriceDown = pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0)
		quantityDown = utils.ConvStrToFloat64(risk.PositionAmt) * -1
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		initPriceUp = pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0)
		quantityUp = utils.ConvStrToFloat64(risk.PositionAmt)
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBackTrading(
			config,
			client,
			pairProcessor,
			tickSizeExp,
			minNotional,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,
		futures.OrderTypeTrailingStopMarket,
		futures.SideTypeBuy,
		futures.OrderTypeTrailingStopMarket,
		quantityUp,
		quantityDown,
		initPriceUp,
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

// Округлення ціни до StepSize знаків після коми
func getStepSizeExp(symbol *futures.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func getTickSizeExp(symbol *futures.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
}

func round(val float64, exp int) float64 {
	return utils.RoundToDecimalPlace(val, exp)
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
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	side futures.SideType,
	grid *grid_types.Grid,
	order *grid_types.Record,
	quantity float64,
	exp int,
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
			upPrice := round(order.GetPrice()*(1+pairProcessor.GetDeltaPrice()), exp)
			if (pair.GetUpBound() == 0 || upPrice <= pair.GetUpBound()) &&
				delta_percent(upPrice) >= config.GetConfigurations().GetPercentsToStopSettingNewOrder() &&
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
					takerRecord = upRecord
					takerOrder = upOrder
				}
			} else {
				if pair.GetUpBound() == 0 || upPrice > pair.GetUpBound() {
					logrus.Debugf("Futures %s: UpBound %v isn't 0 and price %v > UpBound %v",
						pairProcessor.GetPair(), pair.GetUpBound(), upPrice, pair.GetUpBound())
				} else if delta_percent(upPrice) < config.GetConfigurations().GetPercentsToStopSettingNewOrder() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pairProcessor.GetPair(), risk.LiquidationPrice, delta_percent(upPrice), config.GetConfigurations().GetPercentsToStopSettingNewOrder())
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
				takerRecord = downPrice
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair,
				takerOrder.Side,
				grid,
				takerRecord,
				quantity,
				exp,
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
			downPrice := round(order.GetPrice()*(1-pairProcessor.GetDeltaPrice()), exp)
			if (pair.GetLowBound() == 0 || downPrice >= pair.GetLowBound()) &&
				delta_percent(downPrice) >= config.GetConfigurations().GetPercentsToStopSettingNewOrder() &&
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
				if pair.GetLowBound() == 0 || downPrice < pair.GetLowBound() {
					logrus.Debugf("Futures %s: LowBound %v isn't 0 and price %v < LowBound %v",
						pairProcessor.GetPair(), pair.GetLowBound(), downPrice, pair.GetLowBound())
				} else if delta_percent(downPrice) < config.GetConfigurations().GetPercentsToStopSettingNewOrder() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pairProcessor.GetPair(), risk.LiquidationPrice, delta_percent(downPrice), config.GetConfigurations().GetPercentsToStopSettingNewOrder())
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
				config,
				pairProcessor,
				pair,
				takerOrder.Side,
				grid,
				takerRecord,
				quantity,
				exp,
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

func initRun(
	// config *config_types.ConfigFile,
	symbol string,
	client *futures.Client,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	callbackRate float64,
	quit chan struct{}) (pairProcessor *PairProcessor, err error) {
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
		callbackRate,
		quit)
	if err != nil {
		printError()
		return
	}

	if pairProcessor.GetFreeBalance() == 0 {
		printError()
		return
	}
	if marginType != "" && marginType != pairProcessor.GetMarginType() {
		logrus.Debugf("Futures %s set MarginType %v from config into account", pairProcessor.GetPair(), pairProcessor.GetMarginType())
		pairProcessor.SetMarginType(marginType)
	}
	if leverage != 0 && leverage != pairProcessor.GetLeverage() {
		logrus.Debugf("Futures %s set Leverage %v from config into account", pairProcessor.GetPair(), pairProcessor.GetLeverage())
		pairProcessor.SetLeverage(leverage)
	}
	return
}

// func updateConfig(config *config_types.ConfigFile, pair *pairs_types.Pairs) {
// 	if config.GetConfigurations().GetReloadConfig() {
// 		temp := config_types.NewConfigFile(config.GetFileName())
// 		temp.Load()
// 		t_pair := config.GetConfigurations().GetPair(
// 			pair.GetAccountType(),
// 			pair.GetStrategy(),
// 			pair.GetStage(),
// 			pair.GetPair())

// 		pair.SetLimitOnPosition(t_pair.GetLimitOnPosition())
// 		pair.SetLimitOnTransaction(t_pair.GetLimitOnTransaction())
// 		pair.SetDeltaPrice(t_pair.GetDeltaPrice())
// 		pair.SetDeltaQuantity(t_pair.GetDeltaQuantity())
// 		pair.SetUpBound(t_pair.GetUpBound())
// 		pair.SetLowBound(t_pair.GetLowBound())

// 		config.GetConfigurations().SetLogLevel(temp.GetConfigurations().GetLogLevel())
// 		config.GetConfigurations().SetReloadConfig(temp.GetConfigurations().GetReloadConfig())
// 		config.GetConfigurations().SetObservePriceLiquidation(temp.GetConfigurations().GetObservePriceLiquidation())
// 		config.GetConfigurations().SetObservePosition(temp.GetConfigurations().GetObservePosition())
// 		config.GetConfigurations().SetClosePositionOnRestart(temp.GetConfigurations().GetClosePositionOnRestart())
// 		config.GetConfigurations().SetBalancingOfMargin(temp.GetConfigurations().GetBalancingOfMargin())
// 		config.GetConfigurations().SetPercentsToStopSettingNewOrder(temp.GetConfigurations().GetPercentsToStopSettingNewOrder())
// 		config.GetConfigurations().SetPercentToDecreasePosition(temp.GetConfigurations().GetPercentToDecreasePosition())
// 		config.GetConfigurations().SetObserverTimeOutMillisecond(temp.GetConfigurations().GetObserverTimeOutMillisecond())
// 		config.GetConfigurations().SetUsingBreakEvenPrice(temp.GetConfigurations().GetUsingBreakEvenPrice())

// 		config.Save()
// 	}
// }

func getSymbol(
	pairProcessor *PairProcessor) (res *futures.Symbol, err error) {
	val := pairProcessor.GetSymbol()
	if val == nil {
		printError()
		return nil, fmt.Errorf("futures %s: Symbol not found", pairProcessor.GetPair())
	}
	return val.GetFuturesSymbol()
}

func initVars(
	pairProcessor *PairProcessor) (
	symbol *futures.Symbol,
	price float64,
	priceUp,
	priceDown float64,
	minNotional float64,
	tickSizeExp,
	stepSizeExp int,
	err error) {
	symbol, err = getSymbol(pairProcessor)
	if err != nil {
		return
	}
	tickSizeExp = getTickSizeExp(symbol)
	stepSizeExp = getStepSizeExp(symbol)
	// Отримання середньої ціни
	price, _ = pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
	price = round(price, tickSizeExp)
	priceUp = pairProcessor.nextPriceUp(price, 0)
	priceDown = pairProcessor.nextPriceDown(price, 0)
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
	priceDown float64,
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
		priceUp,
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
		priceDown,
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

// func getCurrentPrice(
// 	client *futures.Client,
// 	pairProcessor *PairProcessor,
// 	tickSizeExp int) (currentPrice float64) {
// 	val, _ := pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
// 	currentPrice = round(val, tickSizeExp)
// 	return
// }

func marginBalancing(
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor,
	free float64,
	tickStepSize int) (freeOut float64, err error) {
	// Балансування маржі як треба
	if utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		delta := round(pairProcessor.GetFreeBalance(), tickStepSize) - round(utils.ConvStrToFloat64(risk.IsolatedMargin), tickStepSize)
		if delta != 0 {
			if delta > 0 && delta < free {
				err = pairProcessor.SetPositionMargin(delta, 1)
				logrus.Debugf("Futures %s: IsolatedMargin %v < current position balance %v and we have enough free %v",
					pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance(), free)
			}
		}
		freeOut = pairProcessor.GetFreeBalance()
	} else {
		freeOut = free
	}
	return
}

func initGrid(
	// pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	price float64,
	quantity float64,
	tickSizeExp int,
	sellOrder,
	buyOrder *futures.CreateOrderResponse) (grid *grid_types.Grid, err error) {
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pairProcessor.GetPair())
	grid = grid_types.New()
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, 0, round(price*(1+pairProcessor.GetDeltaPrice()), tickSizeExp), round(price*(1-pairProcessor.GetDeltaPrice()), tickSizeExp), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pairProcessor.GetPair(), price)
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, round(price*(1+pairProcessor.GetDeltaPrice()), tickSizeExp), quantity, 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pairProcessor.GetPair(), round(price*(1+pairProcessor.GetDeltaPrice()), tickSizeExp))
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, round(price*(1-pairProcessor.GetDeltaPrice()), tickSizeExp), quantity, price, 0, types.SideTypeBuy))
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	return
}

func getCallBack_v1(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	tickSizeExp int,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		free         float64
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
					free = pairProcessor.GetFreeBalance()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						grid.Unlock()
						printError()
						close(quit)
						return
					}
					// Балансування маржі як треба
					free, _ = marginBalancing(risk, pairProcessor, free, tickSizeExp)
					if err != nil {
						grid.Unlock()
					}
					err = processOrder(
						config,
						pairProcessor,
						pair,
						event.OrderTradeUpdate.Side,
						grid,
						order,
						quantity,
						tickSizeExp,
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
	config *config_types.ConfigFile,
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
		tickSizeExp   int
		grid          *grid_types.Grid
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо стрім подій
	pairProcessor, err := initRun(
		pair.GetPair(),
		client,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		callbackRate,
		quit)
	if err != nil {
		return err
	}
	_, initPrice, _, _, minNotional, tickSizeExp, _, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pairProcessor.GetFreeBalance(), pairProcessor.GetLimitOnTransaction())
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.InitPositionGrid(10, initPrice)
	if err != nil {
		logrus.Errorf("Can't check position: %v", err)
		close(quit)
		return
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v1(
			config,
			pair,
			pairProcessor,
			grid,
			tickSizeExp,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(
		futures.SideTypeSell,
		futures.OrderTypeLimit,
		futures.SideTypeBuy,
		futures.OrderTypeLimit,
		quantityUp,
		quantityDown,
		initPriceUp,
		initPriceDown,
		pairProcessor)
	if err != nil {
		printError()
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pairProcessor, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
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
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	tickSizeExp int,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		free         float64
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
				free = pairProcessor.GetFreeBalance()
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				// Балансування маржі як треба
				free, _ = marginBalancing(risk, pairProcessor, free, tickSizeExp)
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				err = processOrder(
					config,
					pairProcessor,
					pair,
					event.OrderTradeUpdate.Side,
					grid,
					order,
					quantity,
					tickSizeExp,
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
	config *config_types.ConfigFile,
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
		tickSizeExp   int
		grid          *grid_types.Grid
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо стрім подій
	pairProcessor, err = initRun(
		pair.GetPair(),
		client,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		callbackRate,
		quit)
	if err != nil {
		return err
	}
	_, initPrice, _, _, minNotional, tickSizeExp, _, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current limitOnTransaction %v",
			minNotional, pairProcessor.GetLimitOnTransaction())
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.InitPositionGrid(10, initPrice)
	if err != nil {
		logrus.Errorf("Can't check position: %v", err)
		close(quit)
		return
	}
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v2(
			config,
			pair,
			pairProcessor,
			grid,
			tickSizeExp,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(
		futures.SideTypeSell,
		futures.OrderTypeLimit,
		futures.SideTypeBuy,
		futures.OrderTypeLimit,
		quantityUp,
		quantityDown,
		initPriceUp,
		initPriceDown,
		pairProcessor)
	if err != nil {
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pairProcessor, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}

func getCallBack_v3(
	pairProcessor *PairProcessor,
	tickSizeExp int,
	minNotional float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
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
				free := pairProcessor.GetFreeBalance()
				risk, err := pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					pairProcessor.CancelAllOrders()
					close(quit)
					return
				}
				// Балансування маржі як треба
				marginBalancing(risk, pairProcessor, free, tickSizeExp)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v3(
					pairProcessor.GetCallbackRate(),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					event.OrderTradeUpdate.Side,
					minNotional,
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
	// pair *pairs_types.Pairs,
	callBackRate float64,
	LastExecutedPrice float64,
	AccumulatedFilledQty float64,
	LastExecutedSide futures.SideType,
	minNotional float64,
	pairProcessor *PairProcessor) (err error) {
	var (
		risk         *futures.PositionRisk
		upPrice      float64
		downPrice    float64
		upQuantity   float64
		downQuantity float64
		sellOrder    *futures.CreateOrderResponse
		buyOrder     *futures.CreateOrderResponse
	)
	risk, _ = pairProcessor.GetPositionRisk()
	free := pairProcessor.GetFreeBalance()
	currentPrice, _ := pairProcessor.GetCurrentPrice()
	positionVal := utils.ConvStrToFloat64(risk.PositionAmt) * currentPrice / float64(pairProcessor.GetLeverage())
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
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
					futures.SideTypeBuy,
					futures.OrderTypeTrailingStopMarket,
					upQuantity,
					AccumulatedFilledQty,
					upPrice,
					pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
					pairProcessor)
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
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
					futures.SideTypeBuy,
					futures.OrderTypeTrailingStopMarket,
					upQuantity,
					AccumulatedFilledQty,
					upPrice,
					pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
					pairProcessor)

			}
			if err != nil {
				logrus.Errorf("Can't open position: %v", err)
				printError()
				return
			}
			logrus.Debugf("Futures %s: Sell Quantity Up %v * upPrice %v = %v, minNotional %v, status %v",
				pairProcessor.GetPair(), upQuantity, upPrice, upQuantity*upPrice, minNotional, sellOrder.Status)
			logrus.Debugf("Futures %s: Buy Quantity Down %v * downPrice %v = %v, minNotional %v, status %v",
				pairProcessor.GetPair(),
				AccumulatedFilledQty,
				pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				AccumulatedFilledQty*pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				minNotional,
				buyOrder.Status)
		} else {
			// Створюємо ордер на купівлю, тобто скорочуємо позицію short
			buyOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeBuy,
				futures.OrderTypeTrailingStopMarket,
				AccumulatedFilledQty,
				pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				callBackRate,
				false)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, stopPrice %v, callbackRate %v, status %v",
					pairProcessor.GetPair(),
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
					AccumulatedFilledQty,
					pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
					pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
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
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
					futures.SideTypeBuy,
					futures.OrderTypeTrailingStopMarket,
					AccumulatedFilledQty,
					downQuantity,
					pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
					downPrice,
					pairProcessor)
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
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
					futures.SideTypeBuy,
					futures.OrderTypeTrailingStopMarket,
					AccumulatedFilledQty,
					downQuantity,
					pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
					downPrice,
					pairProcessor)
			}
			if err != nil {
				logrus.Errorf("Can't open position: %v", err)
				printError()
				return
			}
			logrus.Debugf("Futures %s: Sell Quantity Up %v * upPrice %v = %v, minNotional %v, status %v",
				pairProcessor.GetPair(),
				AccumulatedFilledQty,
				pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				AccumulatedFilledQty*pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0),
				minNotional,
				sellOrder.Status)
			logrus.Debugf("Futures %s: Buy Quantity Down %v * downPrice %v = %v, minNotional %v, status %v",
				pairProcessor.GetPair(),
				downQuantity,
				downPrice,
				downPrice*downQuantity,
				minNotional,
				buyOrder.Status)
		} else {
			// Створюємо ордер на продаж, тобто скорочуємо позицію long
			sellOrder, err = createOrder(
				pairProcessor,
				futures.SideTypeSell,
				futures.OrderTypeTrailingStopMarket,
				upQuantity,
				upPrice,
				upPrice,
				callBackRate,
				false)
			if err != nil {
				logrus.Errorf("Futures %s: Could not create Sell order: side %v, type %v, quantity %v, price %v, callbackRate %v, status %v",
					pairProcessor.GetPair(),
					futures.SideTypeSell,
					futures.OrderTypeTrailingStopMarket,
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
		upPrice, upQuantity, downPrice, downQuantity, err = pairProcessor.InitPositionGrid(10, currentPrice)
		if err != nil {
			logrus.Errorf("Can't check position: %v", err)
			return
		}
		logrus.Debugf("Futures %s: Buy Quantity Up %v * upPrice %v = %v, minNotional %v",
			pairProcessor.GetPair(), upQuantity, upPrice, upQuantity*upPrice, minNotional)
		logrus.Debugf("Futures %s: Sell Quantity Down %v * downPrice %v = %v, minNotional %v",
			pairProcessor.GetPair(), downQuantity, downPrice, downQuantity*downPrice, minNotional)
		openPosition(
			futures.SideTypeBuy,
			futures.OrderTypeTrailingStopMarket,
			futures.SideTypeSell,
			futures.OrderTypeTrailingStopMarket,
			upQuantity,
			downQuantity,
			upPrice,
			downPrice,
			pairProcessor)
	}
	return
}

// Працюємо лімітними ордерами (але можливо зменьшувати позицію будемо і TakeProfit ордером),
// відкриваємо ордера на продаж та купівлю з однаковою кількістью
// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
// після 5 наприклад ордера ставимо на більшу відстань
func RunFuturesGridTradingV3(
	config *config_types.ConfigFile,
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
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		free          float64
		initPrice     float64
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		minNotional   float64
		tickSizeExp   int
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо стрім подій
	pairProcessor, err = initRun(
		pair,
		client,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		callbackRate,
		quit)
	if err != nil {
		return err
	}
	free = pairProcessor.GetFreeBalance()
	_, initPrice, _, _, minNotional, tickSizeExp, _, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > free {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v",
			minNotional, free)
	}
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.InitPositionGrid(10, initPrice)
	if err != nil {
		logrus.Errorf("Can't check position: %v", err)
		close(quit)
		return
	}
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		printError()
		close(quit)
		return err
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		initPriceDown = pairProcessor.nextPriceDown(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0)
		quantityDown = pairProcessor.nextQuantityDown(utils.ConvStrToFloat64(risk.PositionAmt)*-1, 0)
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		initPriceUp = pairProcessor.nextPriceUp(utils.ConvStrToFloat64(risk.BreakEvenPrice), 0)
		quantityUp = pairProcessor.nextQuantityUp(utils.ConvStrToFloat64(risk.PositionAmt), 0)
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v3(
			pairProcessor,
			tickSizeExp,
			minNotional,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,
		futures.OrderTypeTrailingStopMarket,
		futures.SideTypeBuy,
		futures.OrderTypeTrailingStopMarket,
		quantityUp,
		quantityDown,
		initPriceUp,
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

func getCallBack_v4(
	client *futures.Client,
	pairProcessor *PairProcessor,
	quantity float64,
	tickSizeExp int,
	stepSizeExp int,
	minNotional float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
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
				free := pairProcessor.GetFreeBalance()
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
				free, _ = marginBalancing(risk, pairProcessor, free, tickSizeExp)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v4(
					client,
					risk,
					minNotional,
					quantity,
					tickSizeExp,
					stepSizeExp,
					free,
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

func createNextPair_v4(
	client *futures.Client,
	risk *futures.PositionRisk,
	minNotional float64,
	quantity float64,
	tickSizeExp int,
	sizeSizeExp int,
	free float64,
	pairProcessor *PairProcessor) (err error) {
	var (
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
	currentPrice, _ := pairProcessor.GetCurrentPrice()
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
	// Визначаємо ціну для нових ордерів
	// Визначаємо кількість для нових ордерів
	upPrice, upQuantity, downPrice, downQuantity, err = pairProcessor.InitPositionGrid(10, currentPrice)
	if err != nil {
		logrus.Errorf("Can't check position: %v", err)
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
			risk, _ = pairProcessor.GetPositionRisk()
			return createNextPair_v4(
				client,
				risk,
				minNotional,
				quantity,
				tickSizeExp,
				sizeSizeExp,
				free,
				pairProcessor)
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
			risk, _ = pairProcessor.GetPositionRisk()
			return createNextPair_v4(
				client,
				risk,
				minNotional,
				quantity,
				tickSizeExp,
				sizeSizeExp,
				free,
				pairProcessor)
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
	config *config_types.ConfigFile,
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
		tickSizeExp   int
		stepSizeExp   int
		pairProcessor *PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо стрім подій
	pairProcessor, err = initRun(
		pair,
		client,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		callbackRate,
		quit)
	if err != nil {
		return err
	}

	_, initPrice, initPriceUp, initPriceDown, minNotional, tickSizeExp, _, err = initVars(pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pairProcessor.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position limitOnTransaction %v",
			minNotional, pairProcessor.GetLimitOnTransaction())
	}
	if config.GetConfigurations().GetDynamicDelta() || config.GetConfigurations().GetDynamicQuantity() {
		initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.InitPositionGrid(10, initPrice)
		if err != nil {
			logrus.Errorf("Can't check position: %v", err)
			close(quit)
			return
		}
		quantity = math.Min(quantityUp, quantityDown)
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v4(
			client,
			pairProcessor,
			quantity,
			tickSizeExp,
			stepSizeExp,
			minNotional,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,
		futures.OrderTypeTrailingStopMarket,
		futures.SideTypeBuy,
		futures.OrderTypeTrailingStopMarket,
		quantityUp,
		quantityDown,
		initPriceUp,
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
				config,
				client,
				pair,
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				wg)

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			err = RunFuturesTrading(
				config,
				pair.GetPair(),
				client,
				degree,
				limit,
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				time.Second,
				debug,
				wg)

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			err = RunFuturesGridTrading(
				config,
				client,
				pair,
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV2 {
			err = RunFuturesGridTradingV2(
				config,
				client,
				pair,
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV3 {
			err = RunFuturesGridTradingV3(
				config,
				client,
				pair.GetPair(),
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV4 {
			err = RunFuturesGridTradingV4(
				config,
				client,
				pair.GetPair(),
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetMarginType(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				quit,
				wg)

			// } else if pair.GetStrategy() == pairs_types.GridStrategyTypeV5 {
			// 	err = RunFuturesGridTradingV5(config, client, pair, quit, wg)

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
