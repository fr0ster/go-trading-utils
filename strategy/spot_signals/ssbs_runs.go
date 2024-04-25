package spot_signals

import (
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	book_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func RunSpotHolding(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	bookTickers := book_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTickers, bookTickerStream.GetDataChannel())

	buyEvent, sellEvent := BuyOrSellSignal(account, bookTickers, pair, stopEvent, triggerEvent)

	collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

	_ = ProcessBuyOrder(
		config, client, account, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, nil, stopEvent)

	<-collectionOutEvent
	pair.SetStage(pairs_types.WorkInPositionStage)
	config.Save()
	stopEvent <- os.Interrupt
	return nil
}

func RunSpotScalping(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	bookTickers := book_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTickers, bookTickerStream.GetDataChannel())

	buyEvent, sellEvent := BuyOrSellSignal(account, bookTickers, pair, stopEvent, triggerEvent)

	_ = ProcessBuyOrder(
		config, client, account, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, nil, stopEvent)

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		workingOutEvent := StopWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)
		_ = ProcessSellOrder(
			config, client, account, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			sellEvent, nil, stopEvent)

		<-workingOutEvent
	}
	return nil
}

func RunSpotTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.TradingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	bookTickers := book_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTickers, bookTickerStream.GetDataChannel())

	buyEvent, sellEvent := BuyOrSellSignal(account, bookTickers, pair, stopEvent, triggerEvent)

	_ = ProcessBuyOrder(
		config, client, account, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, nil, stopEvent)

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.OutputOfPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		orderExecutionGuard := ProcessSellTakeProfitOrder(
			config, client, pair, pairInfo, binance.OrderTypeTakeProfit,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			sellEvent, nil, stopEvent, orderStatusEvent)
		<-orderExecutionGuard
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
	}
	return nil
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.SpotSymbol,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	var (
		stopBuy         = make(chan bool)
		stopSell        = make(chan bool)
		stopProfitOrder = make(chan bool)
	)

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	bookTickers := book_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTickers, bookTickerStream.GetDataChannel())

	buyEvent, sellEvent := BuyOrSellSignal(account, bookTickers, pair, stopEvent, triggerEvent)

	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

			_ = ProcessBuyOrder(
				config, client, account, pair, pairInfo, binance.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				buyEvent, stopBuy, stopEvent)

			<-collectionOutEvent
			pair.SetStage(pairs_types.WorkInPositionStage)
			config.Save()
			stopBuy <- true
			stopEvent <- os.Interrupt
		} else {
			stopEvent <- os.Interrupt
		}

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		_ = ProcessBuyOrder(
			config, client, account, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

			<-collectionOutEvent
			pair.SetStage(pairs_types.WorkInPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			_ = ProcessSellOrder(
				config, client, account, pair, pairInfo, binance.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				sellEvent, stopSell, stopEvent)
		}

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			stopEvent <- os.Interrupt
			return fmt.Errorf("pair %v can't be in WorkInPositionStage for TradingStrategyType", pair.GetPair())
		}
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent, sellEvent)

			_ = ProcessBuyOrder(
				config, client, account, pair, pairInfo, binance.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				buyEvent, stopBuy, stopEvent)

			<-collectionOutEvent
			stopBuy <- true
			pair.SetStage(pairs_types.OutputOfPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.OutputOfPositionStage {
			orderExecutionGuard := ProcessSellTakeProfitOrder(
				config, client, pair, pairInfo, binance.OrderTypeTakeProfit,
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
	}
	return nil
}
