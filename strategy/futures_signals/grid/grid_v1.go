package grid

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func getCallBack_v1(
	// config *config_types.ConfigFile,
	pairProcessor *processor.PairProcessor,
	grid *grid_types.Grid,
	percentsToStopSettingNewOrder items_types.PricePercentType,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     items_types.QuantityType
		locked       items_types.ValueType
		currentPrice items_types.PriceType
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
				currentPrice = items_types.PriceType(utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice))
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
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	upBound items_types.PricePercentType,
	lowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	percentsToStopSettingNewOrder items_types.PricePercentType,
	callbackRate items_types.PricePercentType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		initPrice     items_types.PriceType
		initPriceUp   items_types.PriceType
		initPriceDown items_types.PriceType
		quantity      items_types.QuantityType
		quantityUp    items_types.QuantityType
		quantityDown  items_types.QuantityType
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
	if minNotional > float64(pairProcessor.GetLimitOnTransaction()) {
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
