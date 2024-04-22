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

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

// Виводимо інформацію про позицію
func PositionInfoOut(
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration) {
	for {
		baseBalance, err := account.GetFreeAsset(pair.GetBaseSymbol())
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

func initialization(
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration) (
	depth *depth_types.Depth,
	buyEvent chan *depth_types.DepthItemType,
	sellEvent chan *depth_types.DepthItemType) {
	depth = depth_types.NewDepth(degree, pair.GetPair())

	bookTicker := bookTicker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для отримання оновлення BookTicker через REST
	RestBookTickerUpdater(client, stopEvent, pair, limit, updateTime, bookTicker)
	// Запускаємо потік для отримання оновлення Depth через REST
	RestDepthUpdater(client, stopEvent, pair, limit, updateTime, depth)

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent = BuyOrSellSignal(account, depth, pair, stopEvent, triggerEvent)

	return
}

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
	_, buyEvent, _ :=
		initialization(
			client, degree, limit, pair,
			account, stopEvent, updateTime)

	collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

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
	_, buyEvent, sellEvent :=
		initialization(
			client, degree, limit, pair,
			account, stopEvent, updateTime)

	_ = ProcessBuyOrder(
		config, client, account, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, nil, stopEvent)

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		_ = ProcessSellOrder(
			config, client, account, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			sellEvent, nil, stopEvent)
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
	_, buyEvent, sellEvent :=
		initialization(
			client, degree, limit, pair,
			account, stopEvent, updateTime)

	_ = ProcessBuyOrder(
		config, client, account, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, nil, stopEvent)

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

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
		// depth           *depth_types.Depth
		stopBuy         = make(chan bool)
		stopSell        = make(chan bool)
		stopProfitOrder = make(chan bool)
	)

	baseFree, _ := account.GetFreeAsset(pair.GetBaseSymbol())
	targetFree, _ := account.GetFreeAsset(pair.GetTargetSymbol())

	if pair.GetInitialBalance() == 0 && pair.GetInitialPositionBalance() == 0 {
		pair.SetInitialBalance(baseFree)
		pair.SetInitialPositionBalance(targetFree * pair.GetLimitOnPosition())
		config.Save()
	}

	if pair.GetBuyQuantity() == 0 && pair.GetSellQuantity() == 0 {
		targetFree, err = account.GetFreeAsset(pair.GetPair())
		if err != nil {
			return err
		}
		pair.SetBuyQuantity(targetFree)
		price, err := GetPrice(client, pair.GetPair())
		if err != nil {
			return err
		}
		pair.SetBuyValue(targetFree * price)
		config.Save()
	}

	_, buyEvent, sellEvent :=
		initialization(
			client, degree, limit, pair,
			account, stopEvent, updateTime)

	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

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
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

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
			collectionOutEvent := StartWorkInPositionSignal(account, pair, stopEvent, buyEvent)

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
