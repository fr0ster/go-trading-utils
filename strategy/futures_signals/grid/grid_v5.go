package grid

import (
	"math"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

var (
	v5              sync.Mutex = sync.Mutex{}
	timeOut_v5                 = 1000 * time.Millisecond
	lastResponse_v5            = time.Now()
)

func getCallBack_v5(
	pairProcessor *processor.PairProcessor,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		v5.Lock()
		defer v5.Unlock()
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Type == futures.OrderTypeLimit &&
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
				logrus.Debugf("Futures %s: Risk Position: PositionAmt %v, EntryPrice %v, BreakEvenPrice %v, UnrealizedProfit %v, LiquidationPrice %v, Leverage %v",
					pairProcessor.GetPair(), risk.PositionAmt, risk.EntryPrice, risk.BreakEvenPrice, risk.UnRealizedProfit, risk.LiquidationPrice, risk.Leverage)
				// Балансування маржі як треба
				marginBalancing(risk, pairProcessor)
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
	breakEvenPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
	position := math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))
	if breakEvenPrice == 0 {
		breakEvenPrice = utils.ConvStrToFloat64(risk.EntryPrice)
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
			upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
			upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		}
		downPrice = pairProcessor.NextPriceDown(math.Min(breakEvenPrice, LastExecutedPrice))
		downQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / downPrice)
		if downQuantity > position {
			downQuantity = position
		}
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = true
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		upPrice = pairProcessor.NextPriceUp(math.Max(breakEvenPrice, LastExecutedPrice))
		upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		if upQuantity > position {
			upQuantity = position
		}
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
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
		if upQuantity*upPrice > pairProcessor.GetNotional() {
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
				err = nil // Помилки ігноруємо, якшо не вдалося створити ордер, то чекаємо на перевідкриття
				return
			}
			logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
				pairProcessor.GetPair(), futures.SideTypeSell, futures.OrderTypeLimit, upPrice, upQuantity, pairProcessor.GetCallbackRate(), upOrder.Status)
		}
	} else if LastExecutedSide == futures.SideTypeBuy {
		if downQuantity*downPrice > pairProcessor.GetNotional() {
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
				err = nil // Помилки ігноруємо, якшо не вдалося створити ордер, то чекаємо на перевідкриття
				return
			}
			logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
				pairProcessor.GetPair(), futures.SideTypeBuy, futures.OrderTypeLimit, downPrice, downQuantity, pairProcessor.GetCallbackRate(), downOrder.Status)
		}
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
		logrus.Errorf("Futures %s: %v", pairProcessor.GetPair(), err)
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
	degree int,
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
	targetPercent float64,
	limitDepth depth_types.DepthAPILimit,
	expBase int,
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
		quit,
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
		targetPercent,
		limitDepth,
		expBase,
		callbackRate,
		progression)
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
				if v5.TryLock() {
					logrus.Debugf("Futures %s: Check position", pairProcessor.GetPair())
					free := pairProcessor.GetFreeBalance() * float64(pairProcessor.GetLeverage())
					risk, err := pairProcessor.GetPositionRisk()
					if err != nil {
						printError()
						close(quit)
						return
					}
					currentPrice, err := pairProcessor.GetCurrentPrice()
					if err != nil {
						printError()
						close(quit)
						return
					}
					if pairProcessor.CheckStopLoss(free, risk, currentPrice) {
						logrus.Debugf("Futures %s: Price %v is out of range, close position, LowBound %v, UpBound %v, UnRealizedProfit %v, free %v",
							pairProcessor.GetPair(), currentPrice, pairProcessor.GetLowBound(), pairProcessor.GetUpBound(), risk.UnRealizedProfit, free)
						pairProcessor.CancelAllOrders()
						pairProcessor.ClosePosition(risk)
					}
					// Якщо вже відкрито два ордери,
					// то перевіряємо час їх відкриття та якшо вони відкриті довше 30 хвилин,
					// то закриваємо їх
					if len(openOrders) >= 2 {
						if time.Since(lastResponse_v5) > timeOut_v5*30 {
							logrus.Debugf("Futures %s: Orders are opened too long, cancel all", pairProcessor.GetPair())
							pairProcessor.CancelAllOrders()
						}
					} else if len(openOrders) == 0 {
						// TODO: Перевірити ознаки безпечного входу
						// if pairProcessor.CheckSafeEntry() {
						logrus.Debugf("Futures %s: Open new orders", pairProcessor.GetPair())
						risk, err := pairProcessor.GetPositionRisk()
						if err != nil {
							printError()
							close(quit)
							return
						}
						currentPrice, err := pairProcessor.GetCurrentPrice()
						if err != nil {
							printError()
							close(quit)
							return
						}
						initPosition_v5(currentPrice, risk, pairProcessor, quit)
						lastResponse_v5 = time.Now()
						// }
					}
					v5.Unlock()
				}
			}
		}
	}()
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
