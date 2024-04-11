package spot_signals

import (
	"context"
	"math"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	utils "github.com/fr0ster/go-trading-utils/utils"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"

	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

// Виводимо інформацію про позицію
func PositionInfoOut(
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
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) {
	var (
		depth            *depth_types.Depth
		bookTicker       *bookTicker_types.BookTickerBTree
		stopBuy          = make(chan bool)
		stopSell         = make(chan bool)
		stopByOrSell     = make(chan bool)
		stopAfterProcess = make(chan bool)
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

	// Виводимо інформацію про позицію
	go PositionInfoOut(account, pair, stopEvent, updateTime, price)

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent := BuyOrSellSignal(account, depth, pair, stopEvent, stopByOrSell, bookTickerEvent)

	// Відпрацьовуємо Arbitrage стратегію
	if (*pair).GetStrategy() == pairs_types.ArbitrageStrategyType {
		return

		// Відпрацьовуємо  Holding стратегію
	} else if (*pair).GetStrategy() == pairs_types.HoldingStrategyType {
		if (*pair).GetStage() == pairs_types.InputIntoPositionStage {
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
		}

		// Відпрацьовуємо Scalping стратегію
	} else if (*pair).GetStrategy() == pairs_types.ScalpingStrategyType {
		_ = ProcessBuyOrder(
			config, client, pair, pairInfo, binance.OrderTypeMarket,
			minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
			buyEvent, stopBuy, stopEvent)

		if (*pair).GetStage() != pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

			<-collectionOutEvent
			(*pair).SetStage(pairs_types.WorkInPositionStage)
			config.Save()
		}
		if (*pair).GetStage() != pairs_types.WorkInPositionStage {
			_ = ProcessSellOrder(
				config, client, pair, pairInfo, binance.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				sellEvent, stopSell, stopEvent)
		}

		// Відпрацьовуємо Trading стратегію
	} else if (*pair).GetStrategy() == pairs_types.TradingStrategyType {
		if (*pair).GetStage() != pairs_types.InputIntoPositionStage {
			collectionOutEvent := StartWorkInPositionSignal(account, depth, pair, timeFrame, stopEvent, buyEvent)

			_ = ProcessBuyOrder(
				config, client, pair, pairInfo, binance.OrderTypeMarket,
				minuteOrderLimit, dayOrderLimit, minuteRawRequestLimit,
				buyEvent, stopBuy, stopEvent)

			<-collectionOutEvent
			stopBuy <- true
			(*pair).SetStage(pairs_types.OutputOfPositionStage)
			config.Save()
		}
		if (*pair).GetStage() != pairs_types.OutputOfPositionStage {
			quantityRound := int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).LotSizeFilter().StepSize)))
			priceRound := int(math.Log10(1 / utils.ConvStrToFloat64((*pairInfo).PriceFilter().TickSize)))
			sellQuantity, _ := GetTargetBalance(account, pair)
			order, err :=
				client.NewCreateOrderService().
					Symbol(string(binance.SymbolType((*pair).GetPair()))).
					Type(binance.OrderTypeTakeProfit).
					Side(binance.SideTypeSell).
					Quantity(utils.ConvFloat64ToStr(sellQuantity, quantityRound)).
					Price(utils.ConvFloat64ToStr(price, priceRound)).
					TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
			if err != nil {
				logrus.Errorf("Can't create order: %v", err)
				// logrus.Errorf("Order params: %v", params)
				logrus.Errorf("Symbol: %s, Side: %s, Quantity: %f, Price: %f",
					(*pair).GetPair(), binance.SideTypeSell, sellQuantity, price)
				stopEvent <- os.Interrupt
				return
			}
			ProcessAfterOrder(
				config,
				client,
				pair,
				pairInfo,
				minuteOrderLimit,
				dayOrderLimit,
				minuteRawRequestLimit,
				sellEvent,
				stopAfterProcess,
				stopEvent,
				orderStatusEvent,
				order)
		}

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", (*pair).GetStrategy())
	}
}
