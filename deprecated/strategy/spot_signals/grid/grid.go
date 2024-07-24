package grid

import (
	"math"
	"runtime"
	"sync"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	processor "github.com/fr0ster/go-trading-utils/deprecated/strategy/spot_signals/processor"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"

	utils "github.com/fr0ster/go-trading-utils/utils"
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

func round(val float64, exp int) float64 {
	return utils.RoundToDecimalPlace(float64(val), exp)
}

func initVars(
	pairProcessor *processor.PairProcessor) (
	symbol *binance.Symbol,
	price items_types.PriceType,
	quantity items_types.QuantityType,
	tickSizeExp,
	stepSizeExp int,
	err error) {
	tickSizeExp = pairProcessor.GetTickSizeExp()
	stepSizeExp = pairProcessor.GetStepSizeExp()
	// Отримання середньої ціни
	price, _ = pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
	price = roundPrice(price, symbol)
	setQuantity := func(symbol *binance.Symbol) (quantity items_types.QuantityType) {
		quantity = items_types.QuantityType(round(float64(pairProcessor.GetLimitOnTransaction()/price), stepSizeExp))
		minNotional := utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)
		if float64(quantity)*float64(price) < minNotional {
			quantity = items_types.QuantityType(utils.RoundToDecimalPlace(minNotional/float64(price), stepSizeExp))
		}
		return
	}
	quantity = setQuantity(symbol)
	return
}

func openPosition(
	price items_types.PriceType,
	quantity items_types.QuantityType,
	pairProcessor *processor.PairProcessor) (sellOrder, buyOrder *binance.CreateOrderResponse, err error) {
	var (
		targetBalance items_types.PriceType
	)
	_, _ = pairProcessor.CancelAllOrders()
	// Створюємо ордери на продаж
	if targetBalance, err = pairProcessor.GetTargetBalance(); err == nil && targetBalance >= items_types.PriceType(quantity)*price {
		sellOrder, err = createOrderInGrid(
			pairProcessor,
			binance.SideTypeSell,
			quantity,
			pairProcessor.NextPriceUp(price))
		if err != nil {
			printError()
			return
		}
		logrus.Debugf("Spot %s: Set Sell order on price %v with quantity %v",
			pairProcessor.GetPair(), pairProcessor.NextPriceUp(price), quantity)
	} else {
		logrus.Debugf("Spot %s: Target balance %v >= quantity %v",
			pairProcessor.GetPair(), targetBalance, quantity)
	}
	buyOrder, err = createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, pairProcessor.NextPriceDown(price))
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Spot %s: Set Buy order on price %v with quantity %v",
		pairProcessor.GetPair(), pairProcessor.NextPriceDown(price), quantity)
	return
}

// Створення ордера для розміщення в грід
func createOrderInGrid(
	pairProcessor *processor.PairProcessor,
	side binance.SideType,
	quantity items_types.QuantityType,
	price items_types.PriceType) (order *binance.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		binance.OrderTypeLimit,     // orderType
		side,                       // sideType
		binance.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		0,                          // quantityQty
		price,                      // price
		0,                          // stopPrice
		0)                          // trailingDelta
	return
}

// Округлення ціни до TickSize знаків після коми
func roundPrice(val items_types.PriceType, symbol *binance.Symbol) items_types.PriceType {
	exp := int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
	return items_types.PriceType(utils.RoundToDecimalPlace(float64(val), exp))
}

func getCallBack_v1(
	pairProcessor *processor.PairProcessor,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*binance.WsUserDataEvent) {
	var (
	// quantity float64
	)
	return func(event *binance.WsUserDataEvent) {
		if event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) && !maintainedOrders.Has(grid_types.OrderIdType(event.OrderUpdate.Id)) {
			maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderUpdate.Id))
			logrus.Debugf("Spots %s: Order %v on price %v side %v status %s",
				pairProcessor.GetPair(),
				event.OrderUpdate.Id,
				event.OrderUpdate.Price,
				event.OrderUpdate.Side,
				event.OrderUpdate.Status)

			close(quit)
		}
	}
}

func RunSpotGridTrading(
	client *binance.Client,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpBound items_types.PricePercentType,
	LowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	minSteps int,
	callbackRate items_types.PricePercentType,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity items_types.QuantityType
	)
	// Створюємо обробник пари
	pairProcessor, err := processor.NewPairProcessor(
		stopEvent,
		client,
		symbol,
		limitOnPosition,
		limitOnTransaction,
		UpBound,
		LowBound,
		deltaPrice,
		deltaQuantity,
		callbackRate)
	if err != nil {
		printError()
		return
	}
	_, initPrice, quantity, _, _, err := initVars(pairProcessor)
	if err != nil {
		return err
	}
	go func() {
		for {
			<-stopEvent
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
			return
		}
	}()
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		stopEvent,
		getCallBack_v1(
			pairProcessor,
			stopEvent,
			maintainedOrders),
		nil)
	if err != nil {
		printError()
		return err
	}
	_, _, err = openPosition(initPrice, quantity, pairProcessor)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Spot %s: Start Order Processing", pairProcessor.GetPair())
	<-stopEvent

	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
