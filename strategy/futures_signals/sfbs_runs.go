package futures_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

// Виводимо інформацію про позицію
func PositionInfoOut(
	account account_interfaces.Accounts,
	pair config_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration) {
	for {
		baseBalance, err := account.GetAsset(pair.GetBaseSymbol())
		if err != nil {
			logrus.Errorf("Can't get %s asset: %v", pair.GetBaseSymbol(), err)
			stopEvent <- os.Interrupt
			return
		}
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return
		default:
			if val := pair.GetMiddlePrice(); val != 0 {
				logrus.Infof("Middle %s price: %f, available USDT: %f",
					pair.GetPair(), val, baseBalance)
			}
		}
		time.Sleep(updateTime)
	}
}

func Initialization(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *futures.WsUserDataEvent) (
	depth *depth_types.Depth,
	buyEvent chan *depth_types.DepthItemType,
	sellEvent chan *depth_types.DepthItemType) {
	depth = depth_types.NewDepth(degree, pair.GetPair())

	bookTicker := bookTicker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := futures_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := futures_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	RestUpdate(client, stopEvent, pair, depth, limit, bookTicker, updateTime)

	// Виводимо інформацію про позицію
	go PositionInfoOut(account, pair, stopEvent, updateTime)

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent = BuyOrSellSignal(account, depth, pair, stopEvent, triggerEvent)

	return
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *futures.WsUserDataEvent) {
	var (
		depth           *depth_types.Depth
		stopBuy         = make(chan bool)
		stopSell        = make(chan bool)
		stopProfitOrder = make(chan bool)
	)
	depth, buyEvent, sellEvent :=
		Initialization(
			config, client, degree, limit, pair, pairInfo, account, stopEvent, updateTime,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit, orderStatusEvent)

	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
			stopEvent <- os.Interrupt
			return
		}

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		_ = ProcessBuyOrder(
			config, client, account, pair, pairInfo, futures.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, stopEvent, buyEvent)

			<-collectionOutEvent
			pair.SetStage(pairs_types.WorkInPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			_ = ProcessSellOrder(
				config, client, account, pair, pairInfo, futures.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				sellEvent, stopSell, stopEvent)
		}

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			stopEvent <- os.Interrupt
			return
		}
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, stopEvent, buyEvent)

			_ = ProcessBuyOrder(
				config, client, account, pair, pairInfo, futures.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				buyEvent, stopBuy, stopEvent)

			<-collectionOutEvent
			stopBuy <- true
			pair.SetStage(pairs_types.OutputOfPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.OutputOfPositionStage {
			orderExecutionGuard := ProcessSellTakeProfitOrder(
				config, client, pair, pairInfo, futures.OrderTypeTakeProfit,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				sellEvent, stopProfitOrder, stopEvent, orderStatusEvent)
			<-orderExecutionGuard
			pair.SetStage(pairs_types.PositionClosedStage)
			config.Save()
			stopEvent <- os.Interrupt
		}

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return
	}
}
