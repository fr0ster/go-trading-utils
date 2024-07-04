package futures_signals

import (
	"math"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

var (
	v5         sync.Mutex = sync.Mutex{}
	timeOut_v5            = 1000 * time.Millisecond
)

func getCallBack_v5(
	pairProcessor *processor.PairProcessor,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		v5.Lock()
		defer v5.Unlock()
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
				logrus.Debugf("Futures %s: Risk Position: PositionAmt %v, EntryPrice %v, UnrealizedProfit %v, LiquidationPrice %v, Leverage %v",
					pairProcessor.GetPair(), risk.PositionAmt, risk.EntryPrice, risk.UnRealizedProfit, risk.LiquidationPrice, risk.Leverage)
				// Балансування маржі як треба
				// marginBalancing(risk, pairProcessor)
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v5(
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
					event.OrderTradeUpdate.Side,
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

func getErrorHandling_v5(
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) futures.ErrHandler {
	return func(networkErr error) {
		var (
			initPriceUp    float64
			initPriceDown  float64
			quantityUp     float64
			quantityDown   float64
			reduceOnlyUp   bool
			reduceOnlyDown bool
		)
		openOrders, _ := pairProcessor.GetOpenOrders()
		if len(openOrders) == 0 {
			if v5.TryLock() {
				defer v5.Unlock()
				logrus.Debugf("Futures %s: Error: %v", pairProcessor.GetPair(), networkErr)
				risk, err := pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					pairProcessor.CancelAllOrders()
					close(quit)
					return
				}
				price, err := pairProcessor.GetCurrentPrice()
				if err != nil {
					printError()
					close(quit)
					return
				}
				initPriceUp,
					quantityUp,
					initPriceDown,
					quantityDown,
					reduceOnlyUp,
					reduceOnlyDown,
					err = pairProcessor.GetPrices(price, risk, false)
				if err != nil {
					printError()
					close(quit)
					return
				}
				// Створюємо початкові ордери на продаж та купівлю
				_, _, err = openPosition(
					futures.SideTypeSell,   // sideUp
					futures.OrderTypeLimit, // typeUp
					futures.SideTypeBuy,    // sideDown
					futures.OrderTypeLimit, // typeDown
					false,                  // closePositionUp
					reduceOnlyUp,           // reduceOnlyUp
					false,                  // closePositionDown
					reduceOnlyDown,         // reduceOnlyDown
					quantityUp,             // quantityUp
					quantityDown,           // quantityDown
					initPriceUp,            // priceUp
					initPriceUp,            // stopPriceUp
					initPriceUp,            // activationPriceUp
					initPriceDown,          // priceDown
					initPriceDown,          // stopPriceDown
					initPriceDown,          // activationPriceDown
					pairProcessor)          // pairProcessor
				if err != nil {
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func createNextPair_v5(
	LastExecutedPrice float64,
	LastExecutedSide futures.SideType,
	pairProcessor *processor.PairProcessor) (err error) {
	var (
		risk              *futures.PositionRisk
		upPrice           float64
		downPrice         float64
		upQuantity        float64
		downQuantity      float64
		upClosePosition   bool
		downClosePosition bool
		upReduceOnly      bool
		downReduceOnly    bool
		upOrder           *futures.CreateOrderResponse
		downOrder         *futures.CreateOrderResponse
	)
	risk, _ = pairProcessor.GetPositionRisk()
	free := pairProcessor.GetFreeBalance() * float64(pairProcessor.GetLeverage())
	currentPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
	position := math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))
	if currentPrice == 0 {
		currentPrice = utils.ConvStrToFloat64(risk.EntryPrice)
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		if position*LastExecutedPrice <= free {
			upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
			upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		}
		downPrice = pairProcessor.NextPriceDown(currentPrice)
		downQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / downPrice)
		if downQuantity > position {
			downQuantity = position
		}
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = true
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		upPrice = pairProcessor.NextPriceUp(currentPrice)
		upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		if upQuantity > position {
			upQuantity = position
		}
		if position*LastExecutedPrice <= free {
			downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
			downQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / downPrice)
		}
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = true
		downReduceOnly = false
	} else { // Немає позиції, відкриваємо нову
		upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
		downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
		upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		downQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / downPrice)
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = false
	}
	if LastExecutedSide == futures.SideTypeSell {
		// Створюємо ордер на продаж
		upOrder, err = pairProcessor.CreateOrder(
			futures.OrderTypeLimit,
			futures.SideTypeSell,
			futures.TimeInForceTypeGTC,
			upQuantity,
			upClosePosition,
			upReduceOnly,
			upPrice,
			upPrice,
			upPrice,
			pairProcessor.GetCallbackRate())
		if err != nil {
			logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
				pairProcessor.GetPair(), futures.SideTypeSell, futures.OrderTypeLimit, upPrice, upQuantity, pairProcessor.GetCallbackRate())
			printError()
			return
		}
		logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
			pairProcessor.GetPair(), futures.SideTypeSell, futures.OrderTypeLimit, upPrice, upQuantity, pairProcessor.GetCallbackRate(), upOrder.Status)
	} else if LastExecutedSide == futures.SideTypeBuy {
		// Створюємо ордер на купівлю
		downOrder, err = pairProcessor.CreateOrder(
			futures.OrderTypeLimit,
			futures.SideTypeBuy,
			futures.TimeInForceTypeGTC,
			downQuantity,
			downClosePosition,
			downReduceOnly,
			downPrice,
			downPrice,
			downPrice,
			pairProcessor.GetCallbackRate())
		if err != nil {
			logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
				pairProcessor.GetPair(), futures.SideTypeBuy, futures.OrderTypeLimit, downPrice, downQuantity, pairProcessor.GetCallbackRate())
			printError()
			return
		}
		logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
			pairProcessor.GetPair(), futures.SideTypeBuy, futures.OrderTypeLimit, downPrice, downQuantity, pairProcessor.GetCallbackRate(), downOrder.Status)
	}
	return
}

func initPosition_v5(
	price float64,
	risk *futures.PositionRisk,
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) {
	var (
		initPriceUp    float64
		initPriceDown  float64
		quantityUp     float64
		quantityDown   float64
		reduceOnlyUp   bool
		reduceOnlyDown bool
	)
	initPriceUp,
		quantityUp,
		initPriceDown,
		quantityDown,
		reduceOnlyUp,
		reduceOnlyDown,
		err := pairProcessor.GetPrices(price, risk, false)
	if err != nil {
		printError()
		close(quit)
		return
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,   // sideUp
		futures.OrderTypeLimit, // typeUp
		futures.SideTypeBuy,    // sideDown
		futures.OrderTypeLimit, // typeDown
		false,                  // closePositionUp
		reduceOnlyUp,           // reduceOnlyUp
		false,                  // closePositionDown
		reduceOnlyDown,         // reduceOnlyDown
		quantityUp,             // quantityUp
		quantityDown,           // quantityDown
		initPriceUp,            // priceUp
		initPriceUp,            // stopPriceUp
		initPriceUp,            // activationPriceUp
		initPriceDown,          // priceDown
		initPriceDown,          // stopPriceDown
		initPriceDown,          // activationPriceDown
		pairProcessor)          // pairProcessor
	if err != nil {
		printError()
		close(quit)
		return
	}
}

func RunFuturesGridTradingV5(
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
	minSteps int,
	callbackRate float64,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup,
	timeout ...time.Duration) (err error) {
	var (
		pairProcessor *processor.PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	if len(timeout) > 0 {
		timeOut_v5 = timeout[0]
	}

	// Створюємо обробник пари
	pairProcessor, err = processor.NewPairProcessor(
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
		minSteps,
		callbackRate,
		progression,
		quit)
	if err != nil {
		printError()
		return
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBack_v5(
			pairProcessor,    // pairProcessor
			maintainedOrders, // maintainedOrders
			quit),
		getErrorHandling_v5(
			pairProcessor, // pairProcessor
			quit))         // quit
	if err != nil {
		printError()
		return err
	}
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо горутину для відслідковування виходу ціни за межі диапазону
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quit:
				return
			case <-time.After(timeOut_v5):
				openOrders, _ := pairProcessor.GetOpenOrders()
				if len(openOrders) == 1 {
					if v5.TryLock() {
						free := pairProcessor.GetFreeBalance() * float64(pairProcessor.GetLeverage())
						risk, _ := pairProcessor.GetPositionRisk()
						if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
							currentPrice, err := pairProcessor.GetCurrentPrice()
							if err != nil {
								printError()
								close(quit)
								return
							}
							if (utils.ConvStrToFloat64(risk.PositionAmt) > 0 && currentPrice < pairProcessor.GetLowBound()) ||
								(utils.ConvStrToFloat64(risk.PositionAmt) < 0 && currentPrice > pairProcessor.GetUpBound()) ||
								math.Abs(utils.ConvStrToFloat64(risk.UnRealizedProfit)) > free {
								pairProcessor.ClosePosition(risk)
							}
							initPosition_v5(currentPrice, risk, pairProcessor, quit)
						}
						v5.Unlock()
					}
				} else if len(openOrders) == 2 {
					if v5.TryLock() && time.Since(lastResponse) > timeOut_v5*30 {
						pairProcessor.CancelAllOrders()
						v5.Unlock()
					}
				} else if len(openOrders) == 0 {
					if v5.TryLock() {
						risk, _ := pairProcessor.GetPositionRisk()
						if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
							currentPrice, err := pairProcessor.GetCurrentPrice()
							if err != nil {
								printError()
								close(quit)
								return
							}
							initPosition_v5(currentPrice, risk, pairProcessor, quit)
						}
						v5.Unlock()
					}
				}
			}
		}
	}()
	// risk, err := pairProcessor.GetPositionRisk()
	// if err != nil {
	// 	printError()
	// 	close(quit)
	// 	return
	// }
	// currentPrice, err := pairProcessor.GetCurrentPrice()
	// if err != nil {
	// 	printError()
	// 	close(quit)
	// 	return
	// }
	// initPosition_v5(currentPrice, risk, pairProcessor, quit)
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
