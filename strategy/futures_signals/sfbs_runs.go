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
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func SignalInitialization(
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	account *futures_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration) (
	depth *depth_types.Depth,
	increaseEvent chan *pair_price_types.PairPrice,
	decreaseEvent chan *pair_price_types.PairPrice) {
	depth = depth_types.NewDepth(degree, pair.GetPair())

	bookTicker := bookTicker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := futures_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := futures_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	// Запускаємо потік для контролю ризиків позиції
	RiskSignal(account, pair, stopEvent, triggerEvent)

	// Запускаємо потік для отримання оновлення BookTicker через REST
	RestBookTickerUpdater(client, stopEvent, pair, limit, updateTime, bookTicker)
	// Запускаємо потік для отримання оновлення Depth через REST
	RestDepthUpdater(client, stopEvent, pair, limit, updateTime, depth)

	// Запускаємо потік для отримання сигналів росту та падіння ціни
	increaseEvent, decreaseEvent = PriceSignal(account, depth, pair, stopEvent, triggerEvent)

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
	// var (
	// 	depth           *depth_types.Depth
	// 	stopBuy         = make(chan bool)
	// 	stopSell        = make(chan bool)
	// 	stopProfitOrder = make(chan bool)
	// )

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

	_, _, _ =
		SignalInitialization(
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
		logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return fmt.Errorf("scalping strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			logrus.Warnf("Stage %v is not implemented yet for %v", pair.GetStage(), pair.GetPair())
		}
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			logrus.Warnf("Stage %v is not implemented yet for %v", pair.GetStage(), pair.GetPair())
		}
		if pair.GetStage() == pairs_types.OutputOfPositionStage {
			logrus.Warnf("Stage %v is not implemented yet for %v", pair.GetStage(), pair.GetPair())
		}

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return
	}
	return nil
}
