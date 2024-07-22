package grid

import (
	"math"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

var (
	v4         sync.Mutex = sync.Mutex{}
	timeOut_v4            = 1000 * time.Millisecond
)

func getCallBack_v4(
	pairProcessor *processor.PairProcessor,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		v4.Lock()
		defer v4.Unlock()
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
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v4(
					items_types.PriceType(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
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

func getErrorHandling_v4(
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) futures.ErrHandler {
	return func(networkErr error) {
		var (
			initPriceUp    items_types.PriceType
			initPriceDown  items_types.PriceType
			quantityUp     items_types.QuantityType
			quantityDown   items_types.QuantityType
			reduceOnlyUp   bool
			reduceOnlyDown bool
		)
		openOrders, _ := pairProcessor.GetOpenOrders()
		if len(openOrders) == 0 {
			if v4.TryLock() {
				defer v4.Unlock()
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

func createNextPair_v4(
	LastExecutedPrice items_types.PriceType,
	pairProcessor *processor.PairProcessor) (err error) {
	var (
		risk              *futures.PositionRisk
		upPrice           items_types.PriceType
		downPrice         items_types.PriceType
		upQuantity        items_types.QuantityType
		downQuantity      items_types.QuantityType
		upClosePosition   bool
		downClosePosition bool
		upReduceOnly      bool
		downReduceOnly    bool
	)
	risk, _ = pairProcessor.GetPositionRisk()
	breakEvenPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
	position := items_types.QuantityType(math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
	if breakEvenPrice == 0 {
		breakEvenPrice = utils.ConvStrToFloat64(risk.EntryPrice)
	}
	if items_types.QuantityType(utils.ConvStrToFloat64(risk.PositionAmt)) < 0 {
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
			upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
			upQuantity = pairProcessor.RoundQuantity(
				items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(upPrice)))
		}
		downPrice = pairProcessor.NextPriceDown(items_types.PriceType(math.Min(breakEvenPrice, float64(LastExecutedPrice))))
		downQuantity = pairProcessor.RoundQuantity(
			items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(downPrice)))
		if downQuantity > position {
			downQuantity = position
		}
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = true
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		upPrice = pairProcessor.NextPriceUp(items_types.PriceType(math.Max(breakEvenPrice, float64(LastExecutedPrice))))
		upQuantity = pairProcessor.RoundQuantity(
			items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(upPrice)))
		if upQuantity > position {
			upQuantity = position
		}
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
			downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
			downQuantity = pairProcessor.RoundQuantity(
				items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(downPrice)))
		}
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = true
		downReduceOnly = false
	} else { // Немає позиції, відкриваємо нову
		upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
		downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
		upQuantity = pairProcessor.RoundQuantity(
			items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(upPrice)))
		downQuantity = pairProcessor.RoundQuantity(
			items_types.QuantityType(float64(pairProcessor.GetLimitOnTransaction()) * float64(pairProcessor.GetLeverage()) / float64(downPrice)))
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = false
	}
	// Створюємо ордер на продаж
	// Створюємо ордер на купівлю
	_, _, err = openPosition(
		futures.SideTypeSell,   // sideUp
		futures.OrderTypeLimit, // typeUp
		futures.SideTypeBuy,    // sideDown
		futures.OrderTypeLimit, // typeDown
		upClosePosition,        // closePositionUp
		upReduceOnly,           // reduceOnlyUp
		downClosePosition,      // closePositionDown
		downReduceOnly,         // reduceOnlyDown
		upQuantity,             // quantityUp
		downQuantity,           // quantityDown
		upPrice,                // priceUp
		upPrice,                // stopPriceUp
		upPrice,                // activationPriceUp
		downPrice,              // priceDown
		downPrice,              // stopPriceDown
		downPrice,              // activationPriceDown
		pairProcessor)          // pairProcessor
	if err != nil {
		printError()
		err = nil // Помилки ігноруємо, якшо не вдалося створити ордер, то чекаємо на перевідкриття
		return
	}
	return
}

func initPosition_v4(
	price items_types.PriceType,
	risk *futures.PositionRisk,
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) {
	var (
		initPriceUp    items_types.PriceType
		initPriceDown  items_types.PriceType
		quantityUp     items_types.QuantityType
		quantityDown   items_types.QuantityType
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

func RunFuturesGridTradingV4(
	client *futures.Client,
	degree int,
	pair string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	upBound items_types.PricePercentType,
	lowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	targetPercent items_types.PricePercentType,
	callbackRate items_types.PricePercentType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup,
	depths *depth_types.Depths,
	timeout ...time.Duration) (err error) {
	var (
		pairProcessor *processor.PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	if len(timeout) > 0 {
		timeOut_v4 = timeout[0]
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
		callbackRate,
		progression,
		depths)
	if err != nil {
		printError()
		return
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBack_v4(
			pairProcessor,    // pairProcessor
			maintainedOrders, // maintainedOrders
			quit),
		getErrorHandling_v4(
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
			case <-time.After(timeOut_v4):
				if v4.TryLock() {
					openOrders, _ := pairProcessor.GetOpenOrders()
					free := pairProcessor.GetFreeBalance() * items_types.ValueType(pairProcessor.GetLeverage())
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
							pairProcessor.GetPair(), currentPrice, pairProcessor.GetLowBound(currentPrice), pairProcessor.GetUpBound(currentPrice), risk.UnRealizedProfit, free)
						pairProcessor.CancelAllOrders()
						pairProcessor.ClosePosition(risk)
					}
					if len(openOrders) == 0 {
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
						initPosition_v4(currentPrice, risk, pairProcessor, quit)
					}
					v4.Unlock()
				}
			}
		}
	}()
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
