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
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

func RunHolding(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	timeFrame time.Duration,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	var (
		depth      *depth_types.Depth
		bookTicker *bookTicker_types.BookTickerBTree
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
	}
	config.Save()

	collectionEvent, collectionOutEvent := HoldingSignal(account, depth, pair, timeFrame, stopEvent, bookTickerEvent)

	_ = ProcessBuyOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		collectionEvent, stopEvent, orderStatusEvent)

	go func() {
		for {
			baseBalance, err := account.GetAsset((*pair).GetBaseSymbol())
			if err != nil {
				logrus.Errorf("Can't get %s asset: %v", (*pair).GetBaseSymbol(), err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				return
			default:
				if val := (*pair).GetMiddlePrice(); val != 0 {
					logrus.Infof("Middle %s price: %f, available USDT: %f, Price: %f",
						(*pair).GetPair(), val, baseBalance, price)
				}
			}
			time.Sleep(updateTime)
		}
	}()

	<-collectionOutEvent
	logrus.Infof("Holding strategy is finished")
}

func RunTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	timeFrame time.Duration,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	var (
		depth      *depth_types.Depth
		bookTicker *bookTicker_types.BookTickerBTree
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
	}
	config.Save()

	collectionEvent, collectionOutEvent := TradingInPositionSignal(account, depth, pair, timeFrame, stopEvent, bookTickerEvent)

	ProcessBuyOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		collectionEvent, stopEvent, orderStatusEvent)

	<-collectionOutEvent

	buyEvent, sellEvent := BuyOrSellSignal(account, depth, pair, stopEvent, bookTickerEvent)

	_ = ProcessBuyOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, stopEvent, orderStatusEvent)
	_ = ProcessSellOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		sellEvent, stopEvent, orderStatusEvent)

	go func() {
		for {
			baseBalance, err := account.GetAsset((*pair).GetBaseSymbol())
			if err != nil {
				logrus.Errorf("Can't get %s asset: %v", (*pair).GetBaseSymbol(), err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				return
			default:
				if val := (*pair).GetMiddlePrice(); val != 0 {
					logrus.Infof("Middle %s price: %f, available USDT: %f, Price: %f",
						(*pair).GetPair(), val, baseBalance, price)
				}
			}
			time.Sleep(updateTime)
		}
	}()
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol,
	account account_interfaces.Accounts,
	stopEvent chan os.Signal,
	orderStatusEvent chan *binance.WsUserDataEvent,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	var (
		depth      *depth_types.Depth
		bookTicker *bookTicker_types.BookTickerBTree
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
	}
	config.Save()

	buyEvent, sellEvent := Spot_depth_buy_sell_signals(account, depth, pair, stopEvent, bookTickerEvent)

	ProcessBuyOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		buyEvent, stopEvent, orderStatusEvent)
	ProcessSellOrder(
		config, client, pair, pairInfo, binance.OrderTypeMarket,
		minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
		sellEvent, stopEvent, orderStatusEvent)

	go func() {
		for {
			baseBalance, err := account.GetAsset((*pair).GetBaseSymbol())
			if err != nil {
				logrus.Errorf("Can't get %s asset: %v", (*pair).GetBaseSymbol(), err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				return
			default:
				if val := (*pair).GetMiddlePrice(); val != 0 {
					logrus.Infof("Middle %s price: %f, available USDT: %f, Price: %f",
						(*pair).GetPair(), val, baseBalance, price)
				}
			}
			time.Sleep(updateTime)
		}
	}()
}
