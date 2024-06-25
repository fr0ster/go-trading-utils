package spot_signals

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

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

func RunSpotHolding(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

func RunSpotScalping(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

func RunSpotTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

// Округлення ціни до StepSize знаків після коми
func getStepSizeExp(symbol *binance.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func getTickSizeExp(symbol *binance.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
}

func round(val float64, exp int) float64 {
	return utils.RoundToDecimalPlace(val, exp)
}

func initVars(
	pairProcessor *PairProcessor) (
	symbol *binance.Symbol,
	price,
	quantity float64,
	tickSizeExp,
	stepSizeExp int,
	err error) {
	symbol, err = func() (res *binance.Symbol, err error) {
		val := pairProcessor.GetSymbol()
		if val == nil {
			printError()
			return nil, fmt.Errorf("spot %s: Symbol not found", val.Symbol)
		}
		return val.GetSpotSymbol()
	}()
	if err != nil {
		printError()
		return
	}
	tickSizeExp = getTickSizeExp(symbol)
	stepSizeExp = getStepSizeExp(symbol)
	// Отримання середньої ціни
	price, _ = pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
	price = roundPrice(price, symbol)
	setQuantity := func(symbol *binance.Symbol) (quantity float64) {
		quantity = round(pairProcessor.GetLimitOnTransaction()/price, stepSizeExp)
		minNotional := utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)
		if quantity*price < minNotional {
			quantity = utils.RoundToDecimalPlace(minNotional/price, stepSizeExp)
		}
		return
	}
	quantity = setQuantity(symbol)
	return
}

func openPosition(
	// pair *pairs_types.Pairs,
	price float64,
	quantity float64,
	tickSizeExp int,
	pairProcessor *PairProcessor) (sellOrder, buyOrder *binance.CreateOrderResponse, err error) {
	var (
		targetBalance float64
	)
	_, _ = pairProcessor.CancelAllOrders()
	// Створюємо ордери на продаж
	if targetBalance, err = pairProcessor.GetTargetBalance(); err == nil && targetBalance >= quantity {
		sellOrder, err = createOrderInGrid(
			pairProcessor,
			binance.SideTypeSell,
			quantity,
			pairProcessor.nextPriceUp(price, 0))
		if err != nil {
			printError()
			return
		}
		logrus.Debugf("Spot %s: Set Sell order on price %v with quantity %v",
			pairProcessor.GetPair(), pairProcessor.nextPriceUp(price, 0))
	} else {
		logrus.Debugf("Spot %s: Target balance %v >= quantity %v",
			pairProcessor.GetPair(), targetBalance, quantity)
	}
	buyOrder, err = createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, pairProcessor.nextPriceDown(price, 0))
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Spot %s: Set Buy order on price %v with quantity %v", pairProcessor.GetPair(), pairProcessor.nextPriceDown(price, 0))
	return
}

// Створення ордера для розміщення в грід
func createOrderInGrid(
	pairProcessor *PairProcessor,
	side binance.SideType,
	quantity,
	price float64) (order *binance.CreateOrderResponse, err error) {
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
func roundPrice(val float64, symbol *binance.Symbol) float64 {
	exp := int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
	return utils.RoundToDecimalPlace(val, exp)
}

func createNextPair(
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	currentPrice float64,
	quantity float64,
	limit float64,
	tickSizeExp int) (err error) {
	// Створюємо ордер на продаж
	upPrice := pairProcessor.nextPriceUp(currentPrice, 0)
	if pairProcessor.GetUpBound() != 0 && upPrice <= pairProcessor.GetUpBound() {
		if limit >= quantity {
			_, err = createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, upPrice)
			if err != nil {
				printError()
				return err
			}
			logrus.Debugf("Spots %s: Create Sell order on price %v", pair.GetPair(), upPrice)
		} else {
			logrus.Debugf("Spots %s: Limit %v >= quantity %v or upPrice %v > current position balance %v",
				pair.GetPair(), limit, quantity, upPrice, pairProcessor.GetFreeBalance())
		}
	} else {
		logrus.Debugf("Spots %s: upPrice %v <= upBound %v",
			pair.GetPair(), upPrice, pair.GetUpBound())
	}
	// Створюємо ордер на купівлю
	downPrice := pairProcessor.nextPriceDown(currentPrice, 0)
	if pairProcessor.GetLowBound() != 0 && downPrice >= pairProcessor.GetLowBound() {
		if (limit + quantity*downPrice) <= pairProcessor.GetFreeBalance() {
			_, err = createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, downPrice)
			if err != nil {
				printError()
				return err
			}
			logrus.Debugf("Spots %s: Create Buy order on price %v", pair.GetPair(), downPrice)
		} else {
			logrus.Debugf("Spots %s: Limit %v + quantity %v * downPrice %v <= current position balance %v",
				pair.GetPair(), limit, quantity, downPrice, pairProcessor.GetFreeBalance())
		}
	} else {
		logrus.Debugf("Spots %s: downPrice %v >= lowBound %v",
			pair.GetPair(), downPrice, pair.GetLowBound())
	}
	return nil
}

func getCallBack_v1(
	pairProcessor *PairProcessor,
	tickSizeExp int,
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

		}
	}
}

func RunSpotGridTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	leverage int,
	callbackRate float64,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity float64
	)
	// Створюємо обробник пари
	pairProcessor, err := NewPairProcessor(
		client,
		symbol,
		limitOnPosition,
		limitOnTransaction,
		UpBound,
		LowBound,
		deltaPrice,
		deltaQuantity,
		leverage,
		callbackRate,
		stopEvent,
		false)
	if err != nil {
		printError()
		return
	}
	_, initPrice, quantity, tickSizeExp, _, err := initVars(pairProcessor)
	if err != nil {
		return err
	}
	for {
		<-stopEvent
		pairProcessor.CancelAllOrders()
		logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
		return nil
	}
	maintainedOrders := btree.New(2)
	_, err = pairProcessor.UserDataEventStart(
		getCallBack_v1(
			pairProcessor,
			tickSizeExp,
			stopEvent,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	_, _, err = openPosition(initPrice, quantity, tickSizeExp, pairProcessor)
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

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			logrus.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			logrus.Error(RunSpotHolding(config, client, degree, limit, pair, stopEvent, updateTime, debug, wg))

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			logrus.Error(RunSpotScalping(config, client, degree, limit, pair, stopEvent, updateTime, debug, wg))

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			logrus.Error(RunSpotTrading(config, client, degree, limit, pair, stopEvent, updateTime, debug, wg))

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			logrus.Error(RunSpotGridTrading(
				config,
				client,
				pair.GetPair(),
				pair.GetLimitOnPosition(),
				pair.GetLimitOnTransaction(),
				pair.GetUpBound(),
				pair.GetLowBound(),
				pair.GetDeltaPrice(),
				pair.GetDeltaQuantity(),
				pair.GetLeverage(),
				pair.GetCallbackRate(),
				stopEvent,
				wg))

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			logrus.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
	}()
}
