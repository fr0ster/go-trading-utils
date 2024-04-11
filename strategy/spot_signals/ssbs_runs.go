package spot_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

// Виводимо інформацію про позицію
func positionInfoOut(
	account account_interfaces.Accounts,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	price float64) {
	for {
		baseBalance, err := account.GetAsset((*pair).GetBaseSymbol())
		if err != nil {
			logrus.Errorf("Can't get %s asset: %v", (*pair).GetBaseSymbol(), err)
			stopEvent <- os.Interrupt
			return
		}
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return
		default:
			if val := (*pair).GetMiddlePrice(); val != 0 {
				logrus.Infof("Middle %s price: %f, available USDT: %f, Price: %f",
					(*pair).GetPair(), val, baseBalance, price)
			}
		}
		time.Sleep(updateTime)
	}
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	timeFrame time.Duration,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	var (
		depth        *depth_types.Depth
		bookTicker   *bookTicker_types.BookTickerBTree
		stopBuy      = make(chan bool)
		stopSell     = make(chan bool)
		stopByOrSell = make(chan bool)
	)

	depth = depth_types.NewDepth(degree, (*pair).GetPair())

	bookTicker = bookTicker_types.New(degree)

	_, bookTickerEvent := StartPairStreams((*pair).GetPair(), bookTicker, depth)

	RestUpdate(client, stopEvent, pair, depth, limit, bookTicker, updateTime)

	price, err := GetPrice(client, (*pair).GetPair())
	if err != nil {
		logrus.Errorf("Can't get price: %v", err)
		stopEvent <- os.Interrupt
		return
	}

	if (*pair).GetBuyQuantity() == 0 && (*pair).GetSellQuantity() == 0 {
		targetFree, err := account.GetAsset((*pair).GetTargetSymbol())
		if err != nil {
			logrus.Errorf("Can't get %s asset: %v", (*pair).GetTargetSymbol(), err)
			stopEvent <- os.Interrupt
			return
		}
		(*pair).SetBuyQuantity(targetFree)
		(*pair).SetBuyValue(targetFree * price)
		config.Save()
	}

	// Виводимо інформацію про позицію
	go positionInfoOut(account, pair, stopEvent, updateTime, price)

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent := BuyOrSellSignal(account, depth, pair, stopEvent, stopByOrSell, bookTickerEvent)

	// Відпрацьовуємо Arbitrage стратегію
	if (*pair).GetStrategy() == pairs_types.ArbitrageStrategyType {
		return

		// Відпрацьовуємо  Holding стратегію
	} else if (*pair).GetStrategy() == pairs_types.HoldingStrategyType {
		collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

		_ = ProcessBuyOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		<-collectionOutEvent
		(*pair).SetStage(pairs_types.WorkInPositionStage)
		config.Save()
		stopBuy <- true
		stopByOrSell <- true
		return

		// Відпрацьовуємо Scalping стратегію
	} else if (*pair).GetStrategy() == pairs_types.ScalpingStrategyType {
		collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

		_ = ProcessBuyOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		<-collectionOutEvent
		(*pair).SetStage(pairs_types.WorkInPositionStage)
		config.Save()
		_ = ProcessSellOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			sellEvent, stopSell, stopEvent)

		// positionOutEvent := StartOutputOfPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

		// <-positionOutEvent
		// stopBuy <- true
		// (*pair).SetStage(pairs_types.OutputOfPositionStage)
		// config.Save()

		// StopWorking := StopWorkingSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)
		// <-StopWorking
		// stopSell <- true

		// Відпрацьовуємо Trading стратегію
	} else if (*pair).GetStrategy() == pairs_types.TradingStrategyType {
		collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

		_ = ProcessBuyOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		<-collectionOutEvent
		stopBuy <- true
		(*pair).SetStage(pairs_types.OutputOfPositionStage)
		config.Save()

		_ = ProcessSellOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			sellEvent, stopSell, stopEvent)

		positionOutEvent := StartOutputOfPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

		<-positionOutEvent
		stopSell <- true
		return

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", (*pair).GetStrategy())
	}
}
