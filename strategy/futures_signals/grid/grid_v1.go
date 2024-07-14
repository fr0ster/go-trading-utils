package grid

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func getCallBack_v1(
	// config *config_types.ConfigFile,
	pairProcessor *processor.PairProcessor,
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

func RunFuturesGridTradingV1(
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
	percentsToStopSettingNewOrder float64,
	targetPercent float64,
	limitPercent float64,
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
	pairProcessor, err := processor.NewPairProcessor(
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
		callbackRate,
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
		quit,
		getCallBack_v1(
			pairProcessor,
			grid,
			percentsToStopSettingNewOrder,
			quit,
			maintainedOrders),
		nil)
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
