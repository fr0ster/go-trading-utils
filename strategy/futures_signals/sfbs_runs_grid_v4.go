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

func getCallBack_v4(
	pairProcessor *processor.PairProcessor,
	maintainedOrders *btree.BTree,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	var (
		oldPrice float64
	)
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
				oldPrice, err = createNextPair_v4(
					oldPrice,
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
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
	oldPrice float64,
	LastExecutedPrice float64,
	pairProcessor *processor.PairProcessor) (oldPriceOut float64, err error) {
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
	)
	risk, _ = pairProcessor.GetPositionRisk()
	free := pairProcessor.GetFreeBalance() * float64(pairProcessor.GetLeverage())
	positionVal := utils.ConvStrToFloat64(risk.PositionAmt) * LastExecutedPrice / float64(pairProcessor.GetLeverage())
	oldPriceOut = LastExecutedPrice
	if positionVal < 0 {
		if positionVal >= -free {
			upPrice = pairProcessor.NextPriceUp(oldPrice)
			upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		}
		downPrice = oldPrice
		downQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / downPrice)
		upClosePosition = false
		downClosePosition = false
		upReduceOnly = false
		downReduceOnly = true
	} else if positionVal > 0 {
		upPrice = oldPrice
		upQuantity = pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() * float64(pairProcessor.GetLeverage()) / upPrice)
		if positionVal <= free {
			downPrice = pairProcessor.NextPriceDown(oldPrice)
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
	// Створюємо ордер на продаж, тобто скорочуємо позицію long
	// Створюємо ордер на купівлю, тобто збільшуємо позицію long
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
		return
	}
	return
}

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
	minSteps int,
	callbackRate float64,
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
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBack_v4(
			pairProcessor,    // pairProcessor
			maintainedOrders, // maintainedOrders
			quit))            // quit
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
						initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(currentPrice, risk, false)
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
							false,                  // reduceOnlyUp
							false,                  // closePositionDown
							false,                  // reduceOnlyDown
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
	}()
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
	initPriceUp, quantityUp, initPriceDown, quantityDown, err = pairProcessor.GetPrices(price, risk, false)
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
		false,                  // reduceOnlyUp
		false,                  // closePositionDown
		false,                  // reduceOnlyDown
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
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
