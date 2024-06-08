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

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

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

func checkRun(
	pair *pairs_types.Pairs,
	accountType pairs_types.AccountType,
	strategyType pairs_types.StrategyType) error {
	if pair.GetAccountType() != accountType {
		printError()
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != strategyType {
		printError()
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		printError()
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	return nil
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
	err = checkRun(pair, pairs_types.SpotAccountType, pairs_types.HoldingStrategyType)
	if err != nil {
		return err
	}

	if config.GetConfigurations().GetReloadConfig() {
		go func() {
			for {
				<-time.After(reloadTime)
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
			}
		}()
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	buyEvent, _ := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)

	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-time.After(updateTime):
				triggerEvent <- true
			}
		}
	}()

	collectionOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)

	pairStream, err := NewPairStreams(client, pair, stopEvent, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), stopEvent, debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	_, err = pairProcessor.ProcessBuyOrder(buyEvent)
	if err != nil {
		return err
	}

	<-collectionOutEvent
	pairProcessor.StopBuySignal()
	pair.SetStage(pairs_types.PositionClosedStage)
	config.Save()
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
	err = checkRun(pair, pairs_types.SpotAccountType, pairs_types.ScalpingStrategyType)
	if err != nil {
		return err
	}

	if config.GetConfigurations().GetReloadConfig() {
		go func() {
			for {
				<-time.After(reloadTime)
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
			}
		}()
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, stopEvent, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), stopEvent, debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessSellOrder(sellEvent) // Все одно другий раз не запустится, бо вже працює горутина
		if err != nil {
			return err
		}
		workingOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}

		<-workingOutEvent
		pairProcessor.StopBuySignal()
		pair.SetStage(pairs_types.OutputOfPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		if err != nil {
			return err
		}
		positionClosed := pairObserver.ClosePositionSignal(triggerEvent) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
	}
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
	err = checkRun(pair, pairs_types.SpotAccountType, pairs_types.TradingStrategyType)
	if err != nil {
		return err
	}

	if config.GetConfigurations().GetReloadConfig() {
		go func() {
			for {
				<-time.After(reloadTime)
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
			}
		}()
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, stopEvent, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), stopEvent, debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeTakeProfit) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeTakeProfitLimit)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		pair.SetStage(pairs_types.OutputOfPositionStage) // В trading стратегії не спекулюємо, накопили позицію і закриваемо продажем лімітним ордером
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		// TODO: Закриття позиції лімітним trailing ордером
		quantity, err := GetTargetBalance(account, pair)
		if err != nil {
			return err
		}
		order, err := pairProcessor.CreateOrder(
			binance.OrderTypeTakeProfitLimit,
			binance.SideTypeSell,
			binance.TimeInForceTypeGTC,
			// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
			quantity,
			0,   // quantityQty
			0,   // price
			0,   // stopPrice
			100) // trailingDelta
		if err != nil {
			return err
		}
		positionClosed := pairProcessor.OrderExecutionGuard(order) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
	}
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
func loadConfig(pair *pairs_types.Pairs, config *config_types.ConfigFile, pairStreams *PairStreams) (err error) {
	baseValue, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
	pair.SetCurrentBalance(baseValue)
	config.Save()
	if pair.GetInitialBalance() == 0 {
		pair.SetInitialBalance(baseValue)
		config.Save()
	}
	if pair.GetSellQuantity() == 0 && pair.GetBuyQuantity() == 0 {
		targetValue, err := pairStreams.GetAccount().GetFreeAsset(pair.GetTargetSymbol())
		if err != nil {
			printError()
			return err
		}
		pair.SetBuyQuantity(targetValue)
		config.Save()
	}
	return
}

func initRun(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *pairs_types.Pairs,
	quit chan struct{}) (pairStreams *PairStreams, pairProcessor *PairProcessor, err error) {
	// Створюємо стрім подій
	pairStreams, err = NewPairStreams(client, pair, quit, false)
	if err != nil {
		printError()
		return
	}
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), quit, false)
	if err != nil {
		printError()
		return
	}

	balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
	if err != nil {
		printError()
		return
	}
	pair.SetCurrentBalance(balance)
	config.Save()
	if pair.GetInitialBalance() == 0 {
		pair.SetInitialBalance(balance)
		config.Save()
	}
	return
}

func updateConfig(config *config_types.ConfigFile, pair *pairs_types.Pairs) {
	if config.GetConfigurations().GetReloadConfig() {
		temp := config_types.NewConfigFile(config.GetFileName())
		temp.Load()
		t_pair := config.GetConfigurations().GetPair(
			pair.GetAccountType(),
			pair.GetStrategy(),
			pair.GetStage(),
			pair.GetPair())
		pair.SetLimitOnPosition(t_pair.GetLimitOnPosition())
		pair.SetLimitOnTransaction(t_pair.GetLimitOnTransaction())
		pair.SetSellDelta(t_pair.GetSellDelta())
		pair.SetBuyDelta(t_pair.GetBuyDelta())
		pair.SetUpBound(t_pair.GetUpBound())
		pair.SetLowBound(t_pair.GetLowBound())
	}
}

func initVars(
	client *binance.Client,
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams) (
	symbol *binance.Symbol,
	price,
	quantity float64,
	tickSizeExp,
	stepSizeExp int,
	err error) {
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		err = fmt.Errorf("spot %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
		printError()
		return
	}
	symbol, err = func() (res *binance.Symbol, err error) {
		val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.SpotSymbol{Symbol: pair.GetPair()})
		if val == nil {
			printError()
			return nil, fmt.Errorf("spot %s: Symbol not found", pair.GetPair())
		}
		return val.(*symbol_info.SpotSymbol).GetSpotSymbol()
	}()
	if err != nil {
		printError()
		return
	}
	tickSizeExp = getTickSizeExp(symbol)
	stepSizeExp = getStepSizeExp(symbol)
	// Отримання середньої ціни
	price = roundPrice(pair.GetMiddlePrice(), symbol)
	if price <= 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = roundPrice(price, symbol)
		pair.SetMiddlePrice(price)
		config.Save()
	}
	setQuantity := func(symbol *binance.Symbol) (quantity float64) {
		quantity = round(pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction()/price, stepSizeExp)
		minNotional := utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)
		if quantity*price < minNotional {
			quantity = utils.RoundToDecimalPlace(minNotional/price, stepSizeExp)
		}
		return
	}
	quantity = setQuantity(symbol)
	return
}

func initFirstPairOfOrders(
	pair *pairs_types.Pairs,
	price float64,
	quantity float64,
	tickSizeExp int,
	pairProcessor *PairProcessor) (sellOrder, buyOrder *binance.CreateOrderResponse, err error) {
	_, _ = pairProcessor.CancelAllOrders()
	// Створюємо ордери на продаж
	if pair.GetBuyQuantity()-pair.GetSellQuantity() >= quantity {
		sellOrder, err = createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, round(price*(1+pair.GetSellDelta()), tickSizeExp))
		if err != nil {
			printError()
			return
		}
		logrus.Debugf("Spot %s: Set Sell order on price %v", pair.GetPair(), round(price*(1+pair.GetSellDelta()), tickSizeExp))
	} else {
		logrus.Debugf("Spot %s: BuyQuantity %v - SellQuantity %v >= quantity %v",
			pair.GetPair(), pair.GetBuyQuantity(), pair.GetSellQuantity(), quantity)
	}
	buyOrder, err = createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, round(price*(1-pair.GetBuyDelta()), tickSizeExp))
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Spot %s: Set Buy order on price %v", pair.GetPair(), round(price*(1-pair.GetBuyDelta()), tickSizeExp))
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

func RunSpotGridTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity float64
	)
	err = checkRun(pair, pairs_types.SpotAccountType, pairs_types.GridStrategyType)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairStreams, pairProcessor, err := initRun(config, client, pair, stopEvent)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairStreams)
	if err != nil {
		return err
	}
	_, initPrice, quantity, tickSizeExp, _, err := initVars(client, config, pair, pairStreams)
	if err != nil {
		return err
	}
	_, _, err = initFirstPairOfOrders(pair, initPrice, quantity, tickSizeExp, pairProcessor)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Spot %s: Start Order Processing", pair.GetPair())
	maintainedOrders := btree.New(2)
	for {
		select {
		case <-stopEvent:
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			if event.OrderUpdate.Status == string(binance.OrderStatusTypeFilled) && !maintainedOrders.Has(grid_types.OrderIdType(event.OrderUpdate.Id)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderUpdate.Id))
				logrus.Debugf("Spots %s: Order %v on price %v side %v status %s",
					pair.GetPair(),
					event.OrderUpdate.Id,
					event.OrderUpdate.Price,
					event.OrderUpdate.Side,
					event.OrderUpdate.Status)
				updateConfig(config, pair)
				// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
				if event.OrderUpdate.Side == string(binance.SideTypeSell) {
					pair.SetSellQuantity(pair.GetSellQuantity() + utils.ConvStrToFloat64(event.OrderUpdate.FilledQuoteVolume))
					pair.SetSellValue(pair.GetSellValue() + utils.ConvStrToFloat64(event.OrderUpdate.FilledQuoteVolume)*utils.ConvStrToFloat64(event.OrderUpdate.Price))
				} else if event.OrderUpdate.Side == string(binance.SideTypeBuy) {
					pair.SetBuyQuantity(pair.GetBuyQuantity() + utils.ConvStrToFloat64(event.OrderUpdate.FilledQuoteVolume))
					pair.SetBuyValue(pair.GetBuyValue() + utils.ConvStrToFloat64(event.OrderUpdate.FilledQuoteVolume)*utils.ConvStrToFloat64(event.OrderUpdate.Price))
				}
				err = pair.CalcMiddlePrice()
				if err != nil {
					price, _ := GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
					pair.SetMiddlePrice(round(price, tickSizeExp))
				}
				config.Save()

				pairProcessor.CancelAllOrders()
				logrus.Debugf("Spots %s: Other orders was cancelled", pair.GetPair())
				createNextPair := func(currentPrice float64, quantity float64, limit float64) (err error) {
					// Створюємо ордер на продаж
					upPrice := round(currentPrice*(1+pair.GetSellDelta()), tickSizeExp)
					if limit >= quantity {
						_, err = createOrderInGrid(pairProcessor, binance.SideTypeSell, quantity, upPrice)
						if err != nil {
							printError()
							return err
						}
						logrus.Debugf("Spots %s: Create Sell order on price %v", pair.GetPair(), upPrice)
					} else {
						logrus.Debugf("Spots %s: Limit %v >= quantity %v or upPrice %v > current position balance %v",
							pair.GetPair(), limit, quantity, upPrice, pair.GetCurrentPositionBalance())
					}
					// Створюємо ордер на купівлю
					downPrice := round(currentPrice*(1-pair.GetBuyDelta()), tickSizeExp)
					if (limit + quantity*downPrice) <= pair.GetCurrentPositionBalance() {
						_, err = createOrderInGrid(pairProcessor, binance.SideTypeBuy, quantity, downPrice)
						if err != nil {
							printError()
							return err
						}
						logrus.Debugf("Spots %s: Create Buy order on price %v", pair.GetPair(), downPrice)
					} else {
						logrus.Debugf("Spots %s: Limit %v + quantity %v * downPrice %v <= current position balance %v",
							pair.GetPair(), limit, quantity, downPrice, pair.GetCurrentPositionBalance())
					}
					return nil
				}
				createNextPair(pair.GetMiddlePrice(), quantity, pair.GetBuyQuantity()-pair.GetSellQuantity())
			}
		}
	}
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
			logrus.Error(RunSpotGridTrading(config, client, pair, stopEvent, wg))

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			logrus.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
	}()
}
