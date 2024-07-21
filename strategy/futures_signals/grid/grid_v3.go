package grid

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

var (
	v3         sync.Mutex = sync.Mutex{}
	timeOut_v3            = 1000 * time.Millisecond
)

func getCallBack_v3(
	pairProcessor *processor.PairProcessor,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		v3.Lock()
		defer v3.Unlock()
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
				logrus.Debugf("Futures %s: Risk Position: PositionAmt %v, EntryPrice %v, BreakEvenPrice %v, UnrealizedProfit %v, LiquidationPrice %v, Leverage %v",
					pairProcessor.GetPair(), risk.PositionAmt, risk.EntryPrice, risk.BreakEvenPrice, risk.UnRealizedProfit, risk.LiquidationPrice, risk.Leverage)
				// Балансування маржі як треба
				marginBalancing(risk, pairProcessor)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pairProcessor.GetPair())
				err = createNextPair_v3(
					types.PriceType(utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice)),
					types.QuantityType(utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty)),
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

func getErrorHandling_v3(
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) futures.ErrHandler {
	return func(networkErr error) {
		var (
			initPriceUp    types.PriceType
			initPriceDown  types.PriceType
			quantityUp     types.QuantityType
			quantityDown   types.QuantityType
			reduceOnlyUp   bool
			reduceOnlyDown bool
		)
		openOrders, _ := pairProcessor.GetOpenOrders()
		if len(openOrders) == 0 {
			if v3.TryLock() {
				defer v3.Unlock()
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
					err = pairProcessor.GetPrices(price, risk, true)
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

func createNextPair_v3(
	LastExecutedPrice types.PriceType,
	AccumulatedFilledQty types.QuantityType,
	LastExecutedSide futures.SideType,
	pairProcessor *processor.PairProcessor) (err error) {
	var (
		risk              *futures.PositionRisk
		upPrice           types.PriceType
		downPrice         types.PriceType
		upQuantity        types.QuantityType
		downQuantity      types.QuantityType
		upClosePosition   bool
		downClosePosition bool
		upReduceOnly      bool
		downReduceOnly    bool
	)
	riskBreakEvenPriceOrEntryPrice := func(risk *futures.PositionRisk) types.PriceType {
		breakEvenPrice := types.PriceType(utils.ConvStrToFloat64(risk.BreakEvenPrice))
		entryPrice := types.PriceType(utils.ConvStrToFloat64(risk.EntryPrice))
		if breakEvenPrice != 0 {
			return breakEvenPrice
		} else if entryPrice != 0 {
			return entryPrice
		}
		return 0
	}
	risk, _ = pairProcessor.GetPositionRisk()
	position := types.QuantityType(math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 { // Маємо позицію short
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
			// Виконаний ордер був на продаж, тобто збільшив або відкрив позицію short
			if LastExecutedSide == futures.SideTypeSell {
				// Перевіряємо чи маємо ми записи для розрахунку цінових позицій short
				// Як не маємо, то вважаемо, шо виконаний ордер створив позицію short
				// та розраховуємо цінові позиції від ціни відкриття позиції
				if pairProcessor.GetUpLength() == 0 {
					_, _, _, _, _, err = pairProcessor.InitPositionGridUp(LastExecutedPrice)
					if err != nil {
						err = fmt.Errorf("can't init position up: %v", err)
						printError()
						return
					}
				}
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				upPrice, upQuantity, err = pairProcessor.NextUp(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position up: %v", err)
					printError()
					return
				}
				// Виконаний ордер був на купівлю, тобто скоротив позицію short
				// Обробляємо розворот курсу
			} else if LastExecutedSide == futures.SideTypeBuy {
				logrus.Debugf("Futures %s: ComeBack Price, LastExecutedPrice %v, AccumulatedFilledQty %v",
					pairProcessor.GetPair(), LastExecutedPrice, AccumulatedFilledQty)
				// У випадку, коли ми маємо позицію short,
				// але не маємо розрахованих цінових позицій від ціни відкриття позиції
				// то ми їх розраховуємо
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				upPrice, upQuantity, err = pairProcessor.NextDown(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
					upQuantity = pairProcessor.NextQuantityUp(AccumulatedFilledQty)
				}
			}
			// Створюємо ордер на продаж, тобто збільшуємо позицію short
			// Створюємо ордер на купівлю, тобто скорочуємо позицію short
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			downQuantity = types.QuantityType(math.Min(float64(AccumulatedFilledQty), math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))))
		} else {
			// Створюємо ордер на купівлю, тобто скорочуємо позицію short
			upPrice = pairProcessor.NextPriceUp(riskBreakEvenPriceOrEntryPrice(risk))
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			upQuantity = 0
			downQuantity = types.QuantityType(math.Min(float64(AccumulatedFilledQty), math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))))
		}
		if downQuantity > position {
			downQuantity = position
		}
		// Позиція short, не закриваємо обов'язково повністю ордером на купівлю але скорочуемо
		upClosePosition = false
		upReduceOnly = false
		downClosePosition = false
		downReduceOnly = true
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 { // Маємо позицію long
		if pairProcessor.CheckAddPosition(risk, LastExecutedPrice) {
			// Виконаний ордер був на купівлю, тобто збільшив позицію long
			if LastExecutedSide == futures.SideTypeBuy {
				// Перевіряємо чи маємо ми записи для розрахунку цінових позицій long
				// Як не маємо, то вважаемо, шо виконаний ордер створив позицію long
				// та розраховуємо цінові позиції від ціни відкриття позиції
				if pairProcessor.GetDownLength() == 0 {
					_, _, _, _, _, err = pairProcessor.InitPositionGridDown(LastExecutedPrice)
					if err != nil {
						err = fmt.Errorf("can't init position down: %v", err)
						printError()
						return
					}
				}
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				downPrice, downQuantity, err = pairProcessor.NextDown(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					logrus.Errorf("Can't check position down: %v", err)
					printError()
					return
				}
				// Виконаний ордер був на продаж, тобто скоротив позицію long
				// Обробляємо розворот курсу
			} else if LastExecutedSide == futures.SideTypeSell {
				logrus.Debugf("Futures %s: ComeBack Price, LastExecutedPrice %v, AccumulatedFilledQty %v",
					pairProcessor.GetPair(), LastExecutedPrice, AccumulatedFilledQty)
				// У випадку, коли ми маємо позицію long,
				// але не маємо розрахованих цінових позицій від ціни відкриття позиції
				// то ми їх розраховуємо
				// Визначаємо ціну для нових ордерів
				// Визначаємо кількість для нових ордерів
				downPrice, downQuantity, err = pairProcessor.NextUp(LastExecutedPrice, AccumulatedFilledQty)
				if err != nil {
					downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
					downQuantity = pairProcessor.NextQuantityDown(AccumulatedFilledQty)
				}
			}
			// Створюємо ордер на продаж, тобто скорочуємо позицію long
			// Створюємо ордер на купівлю, тобто збільшуємо позицію long
			upPrice = pairProcessor.NextPriceUp(riskBreakEvenPriceOrEntryPrice(risk))
			upQuantity = types.QuantityType(math.Min(float64(AccumulatedFilledQty), math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))))
		} else {
			// Створюємо ордер на продаж, тобто скорочуємо позицію long
			upPrice = pairProcessor.NextPriceUp(riskBreakEvenPriceOrEntryPrice(risk))
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			upQuantity = types.QuantityType(math.Min(float64(AccumulatedFilledQty), math.Abs(utils.ConvStrToFloat64(risk.PositionAmt))))
			downQuantity = 0
		}
		if upQuantity > position {
			upQuantity = position
		}
		// Позиція long, не закриваємо обов'язково повністю ордером на продаж але скорочуемо
		upClosePosition = false
		upReduceOnly = true
		downClosePosition = false
		downReduceOnly = false
	} else { // Немає позиції, відкриваємо нову
		// Відкриваємо нову позицію
		// Визначаємо ціну для нових ордерів
		// Визначаємо кількість для нових ордерів
		pairProcessor.UpDownClear()
		pairProcessor.SetBounds(LastExecutedPrice)
		upPrice = pairProcessor.NextPriceUp(LastExecutedPrice)
		downPrice = pairProcessor.NextPriceDown(LastExecutedPrice)
		_, _, _, upQuantity, _, err = pairProcessor.CalculateInitialPosition(LastExecutedPrice, pairProcessor.UpBound)
		if err != nil {
			logrus.Errorf("Future %s: can't calculate initial position for price up %v", pairProcessor.GetPair(), LastExecutedPrice)
		}
		_, _, _, downQuantity, _, err = pairProcessor.CalculateInitialPosition(LastExecutedPrice, pairProcessor.LowBound)
		if err != nil {
			logrus.Errorf("Future %s: can't calculate initial position for price down %v", pairProcessor.GetPair(), LastExecutedPrice)
		}
		// Тіко відкриваємо позицію, не закриваємо та не скорочуемо
		upClosePosition = false
		upReduceOnly = false
		downClosePosition = false
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

func initPosition_v3(
	price types.PriceType,
	risk *futures.PositionRisk,
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) {
	var (
		initPriceUp    types.PriceType
		initPriceDown  types.PriceType
		quantityUp     types.QuantityType
		quantityDown   types.QuantityType
		reduceOnlyUp   bool
		reduceOnlyDown bool
	)
	initPriceUp,
		quantityUp,
		initPriceDown,
		quantityDown,
		reduceOnlyUp,
		reduceOnlyDown,
		err := pairProcessor.GetPrices(price, risk, true)
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
	minSteps int,
	targetPercent float64,
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
		timeOut_v3 = timeout[0]
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
		progression)
	if err != nil {
		printError()
		return
	}
	// upNewOrder := upPositionNewOrderType
	// downNewOrder := downPositionNewOrderType
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBack_v3(
			pairProcessor,    // pairProcessor
			maintainedOrders, // maintainedOrders
			quit),
		getErrorHandling_v3(
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
			case <-time.After(timeOut_v3):
				if v3.TryLock() {
					openOrders, _ := pairProcessor.GetOpenOrders()
					if len(openOrders) == 1 {
						free := pairProcessor.GetFreeBalance() * types.PriceType(pairProcessor.GetLeverage())
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
								types.PriceType(math.Abs(utils.ConvStrToFloat64(risk.UnRealizedProfit))) > free {
								logrus.Debugf("Futures %s: Price %v is out of range, close position, LowBound %v, UpBound %v, UnRealizedProfit %v, free %v",
									pairProcessor.GetPair(), currentPrice, pairProcessor.GetLowBound(), pairProcessor.GetUpBound(), risk.UnRealizedProfit, free)
								pairProcessor.ClosePosition(risk)
							}
							initPosition_v3(currentPrice, risk, pairProcessor, quit)
						}
					} else if len(openOrders) == 0 {
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
						initPosition_v3(currentPrice, risk, pairProcessor, quit)
					}
					v3.Unlock()
				}
			}
		}
	}()
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
