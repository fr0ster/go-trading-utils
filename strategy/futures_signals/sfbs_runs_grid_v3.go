package futures_signals

import (
	"fmt"
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

func getCallBack_v3(
	pairProcessor *processor.PairProcessor,
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
	pairProcessor *processor.PairProcessor) (err error) {
	var (
		risk              *futures.PositionRisk
		upPrice           float64
		downPrice         float64
		upQuantity        float64
		downQuantity      float64
		upType            futures.OrderType
		downType          futures.OrderType
		upClosePosition   bool
		downClosePosition bool
		upReduceOnly      bool
		downReduceOnly    bool
	)
	riskBreakEvenPriceOrEntryPrice := func(risk *futures.PositionRisk) float64 {
		breakEvenPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
		entryPrice := utils.ConvStrToFloat64(risk.EntryPrice)
		if breakEvenPrice != 0 {
			return breakEvenPrice
		} else if entryPrice != 0 {
			return entryPrice
		}
		return 0
	}
	risk, _ = pairProcessor.GetPositionRisk()
	free := pairProcessor.GetFreeBalance() * float64(pairProcessor.GetLeverage())
	currentPrice, _ := pairProcessor.GetCurrentPrice()
	positionVal := utils.ConvStrToFloat64(risk.PositionAmt) * currentPrice / float64(pairProcessor.GetLeverage())
	if positionVal < 0 { // Маємо позицію short
		if positionVal >= -free {
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
			upType = shortPositionIncOrderType
			downType = shortPositionDecOrderType
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			downQuantity = math.Min(AccumulatedFilledQty, math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
		} else {
			// Створюємо ордер на купівлю, тобто скорочуємо позицію short
			upType = shortPositionNewOrderType // На справді кількість буде нульова але тип ордера має бути вказаний
			downType = shortPositionNewOrderType
			upPrice = pairProcessor.NextPriceUp(riskBreakEvenPriceOrEntryPrice(risk))
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			upQuantity = 0
			downQuantity = math.Min(AccumulatedFilledQty, math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
		}
		// Позиція short, не закриваємо обов'язково повністю ордером на купівлю але скорочуемо
		upClosePosition = false
		upReduceOnly = false
		downClosePosition = false
		downReduceOnly = true
	} else if positionVal > 0 { // Маємо позицію long
		if positionVal <= free {
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
			upType = longPositionDecOrderType
			downType = longPositionIncOrderType
			upQuantity = math.Min(AccumulatedFilledQty, math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
		} else {
			// Створюємо ордер на продаж, тобто скорочуємо позицію long
			upPrice = pairProcessor.NextPriceUp(riskBreakEvenPriceOrEntryPrice(risk))
			downPrice = pairProcessor.NextPriceDown(riskBreakEvenPriceOrEntryPrice(risk))
			upType = longPositionDecOrderType
			downType = longPositionDecOrderType // На справді кількість буде нульова але тип ордера має бути вказаний
			upQuantity = math.Min(AccumulatedFilledQty, math.Abs(utils.ConvStrToFloat64(risk.PositionAmt)))
			downQuantity = 0
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
		upType = shortPositionNewOrderType
		downType = longPositionNewOrderType
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
	// Створюємо ордер на продаж, тобто скорочуємо позицію long
	// Створюємо ордер на купівлю, тобто збільшуємо позицію long
	_, _, err = openPosition(
		futures.SideTypeSell, // sideUp
		upType,               // typeUp
		futures.SideTypeBuy,  // sideDown
		downType,             // typeDown
		upClosePosition,      // closePositionUp
		upReduceOnly,         // reduceOnlyUp
		downClosePosition,    // closePositionDown
		downReduceOnly,       // reduceOnlyDown
		upQuantity,           // quantityUp
		downQuantity,         // quantityDown
		upPrice,              // priceUp
		upPrice,              // stopPriceUp
		upPrice,              // activationPriceUp
		downPrice,            // priceDown
		downPrice,            // stopPriceDown
		downPrice,            // activationPriceDown
		pairProcessor)        // pairProcessor
	if err != nil {
		printError()
		return
	}
	return
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
	callbackRate float64,
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionIncOrderType futures.OrderType,
	shortPositionDecOrderType futures.OrderType,
	longPositionIncOrderType futures.OrderType,
	longPositionDecOrderType futures.OrderType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup,
	timeout ...time.Duration) (err error) {
	var (
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		pairProcessor *processor.PairProcessor
		timeOut       time.Duration = 5000
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	if len(timeout) > 0 {
		timeOut = timeout[0]
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
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		printError()
		return err
	}
	price, err := pairProcessor.GetCurrentPrice()
	if err != nil {
		printError()
		return err
	}
	upNewOrder := upPositionNewOrderType
	downNewOrder := downPositionNewOrderType
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, risk, true)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBack_v3(
			pairProcessor,             // pairProcessor
			upPositionNewOrderType,    // shortPositionNewOrderType,
			shortPositionIncOrderType, // shortPositionIncOrderType,
			shortPositionDecOrderType, // shortPositionDecOrderType,
			downPositionNewOrderType,  // longPositionNewOrderType,
			longPositionIncOrderType,  // longPositionIncOrderType,
			longPositionDecOrderType,  // longPositionDecOrderType,
			maintainedOrders,          // maintainedOrders
			quit))                     // quit
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
			case <-time.After(timeOut * time.Millisecond):
				openOrders, _ := pairProcessor.GetOpenOrders()
				if len(openOrders) == 1 {
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
						initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, risk, true)
						if err != nil {
							printError()
							close(quit)
							return
						}
						// Створюємо початкові ордери на продаж та купівлю
						_, _, err = openPosition(
							upOrderSideOpen,   // sideUp
							upNewOrder,        // typeUp
							downOrderSideOpen, // sideDown
							downNewOrder,      // typeDown
							false,             // closePositionUp
							false,             // reduceOnlyUp
							false,             // closePositionDown
							false,             // reduceOnlyDown
							quantityUp,        // quantityUp
							quantityDown,      // quantityDown
							initPriceUp,       // priceUp
							initPriceUp,       // stopPriceUp
							initPriceUp,       // activationPriceUp
							initPriceDown,     // priceDown
							initPriceDown,     // stopPriceDown
							initPriceDown,     // activationPriceDown
							pairProcessor)     // pairProcessor
						if err != nil {
							printError()
							close(quit)
							return
						}
					}
				}
			}
		}
	}()
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		upOrderSideOpen,   // sideUp
		upNewOrder,        // typeUp
		downOrderSideOpen, // sideDown
		downNewOrder,      // typeDown
		false,             // closePositionUp
		false,             // reduceOnlyUp
		false,             // closePositionDown
		false,             // reduceOnlyDown
		quantityUp,        // quantityUp
		quantityDown,      // quantityDown
		initPriceUp,       // priceUp
		initPriceUp,       // stopPriceUp
		initPriceUp,       // activationPriceUp
		initPriceDown,     // priceDown
		initPriceDown,     // stopPriceDown
		initPriceDown,     // activationPriceDown
		pairProcessor)     // pairProcessor
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
