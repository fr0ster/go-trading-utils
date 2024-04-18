package futures_signals

import (
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

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
	account *futures_account.Account,
	pair pairs_interfaces.Pairs,
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

func initialization(
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	account *futures_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration) (
	depth *depth_types.Depth,
	buyEvent chan *depth_types.DepthItemType,
	sellEvent chan *depth_types.DepthItemType) {
	depth = depth_types.NewDepth(degree, pair.GetPair())

	bookTicker := bookTicker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := futures_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := futures_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для отримання оновлення BookTicker через REST
	RestBookTickerUpdater(client, stopEvent, pair, limit, updateTime, bookTicker)
	// Запускаємо потік для отримання оновлення Depth через REST
	RestDepthUpdater(client, stopEvent, pair, limit, updateTime, depth)

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent = BuyOrSellSignal(account, depth, pair, stopEvent, triggerEvent)

	return
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	pairInfo *symbol_info_types.FuturesSymbol,
	account *futures_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *futures.WsUserDataEvent) (err error) {
	var (
		depth           *depth_types.Depth
		stopBuy         = make(chan bool)
		stopSell        = make(chan bool)
		stopProfitOrder = make(chan bool)
	)

	baseFree, _ := account.GetAsset(pair.GetBaseSymbol())
	targetFree, _ := account.GetAsset(pair.GetTargetSymbol())

	if pair.GetInitialBalance() == 0 && pair.GetInitialPositionBalance() == 0 {
		pair.SetInitialBalance(baseFree)
		pair.SetInitialPositionBalance(targetFree * pair.GetLimitOnPosition())
		config.Save()
	}

	if pair.GetBuyQuantity() == 0 && pair.GetSellQuantity() == 0 {
		targetFree, err = account.GetAsset(pair.GetPair())
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

	depth, buyEvent, sellEvent :=
		initialization(
			client, degree, limit, pair,
			account, stopEvent, updateTime)

	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
			stopEvent <- os.Interrupt
			return fmt.Errorf("holding strategy should not be used for %v", pair.GetPair())
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
			return fmt.Errorf("pair %v can't be in WorkInPositionStage for TradingStrategyType", pair.GetPair())
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
	return nil
}
