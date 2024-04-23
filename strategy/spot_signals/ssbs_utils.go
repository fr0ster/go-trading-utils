package spot_signals

import (
	"context"
	"errors"
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_bookticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
)

const (
	errorMsg = "Error: %v"
)

func LimitRead(degree int, symbols []string, client *binance.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	exchangeInfo := exchange_types.New()
	spot_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}

func RestBookTickerUpdater(
	client *binance.Client,
	stop chan os.Signal,
	pair pairs_interfaces.Pairs,
	limit int,
	updateTime time.Duration,
	bookTicker *bookTicker_types.BookTickerBTree) {
	go func() {
		for {
			select {
			case <-stop:
				// Якщо отримано сигнал з каналу stop, вийти з циклу
				return
			default:
				err := spot_bookticker.Init(bookTicker, pair.GetPair(), client)
				if err != nil {
					logrus.Errorf(errorMsg, err)
					stop <- os.Interrupt
					return
				}

				time.Sleep(updateTime)
			}
		}
	}()
}

func RestDepthUpdater(
	client *binance.Client,
	stop chan os.Signal,
	pair pairs_interfaces.Pairs,
	limit int,
	updateTime time.Duration,
	depth *depth_types.Depth) {
	go func() {
		for {
			select {
			case <-stop:
				// Якщо отримано сигнал з каналу stop, вийти з циклу
				return
			default:
				err := spot_depth.Init(depth, client, limit)
				if err != nil {
					logrus.Errorf(errorMsg, err)
					stop <- os.Interrupt
					return
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func GetPrice(client *binance.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func GetAskAndBid(depths *depth_types.Depth) (ask float64, bid float64, err error) {
	getPrice := func(val btree.Item) (float64, error) {
		if val == nil {
			return 0, errors.New("value is nil")
		}
		return val.(*depth_types.DepthItemType).Price, nil
	}
	ask, err = getPrice(depths.GetAsks().Min())
	if err != nil {
		return 0, 0, fmt.Errorf("value is nil, can't get ask: %v", err)
	}
	bid, err = getPrice(depths.GetBids().Max())
	if err != nil {
		return 0, 0, fmt.Errorf("value is nil, can't get bid: %v", err)
	}
	return
}

func GetBound(pair pairs_interfaces.Pairs) (boundAsk float64, boundBid float64, err error) {
	boundAsk = pair.GetMiddlePrice() * (1 - pair.GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	boundBid = pair.GetMiddlePrice() * (1 + pair.GetSellDelta())
	logrus.Debugf("Bid bound: %f", boundBid)
	return
}

func GetBaseBalance(
	account account_interfaces.Accounts,
	pair pairs_interfaces.Pairs) (
	baseBalance float64, // Кількість базової валюти
	err error) {
	baseBalance, err = func(pair pairs_interfaces.Pairs) (
		baseBalance float64,
		err error) {
		baseBalance, err = account.GetFreeAsset(pair.GetBaseSymbol())
		return
	}(pair)

	if err != nil {
		return 0, err
	}
	return
}

func GetTargetBalance(
	account account_interfaces.Accounts,
	pair pairs_interfaces.Pairs) (
	targetBalance float64, // Кількість торгової валюти
	err error) {
	targetBalance, err = func(pair pairs_interfaces.Pairs) (
		targetBalance float64,
		err error) {
		targetBalance, err = account.GetFreeAsset(pair.GetTargetSymbol())
		return
	}(pair)

	if err != nil {
		return 0, err
	}
	return
}

func GetCommission(
	account account_interfaces.Accounts) (
	commission float64) { // Комісія за покупку/Комісія за продаж)
	return account.GetMakerCommission()
}

func GetTransactionValue(
	pair pairs_interfaces.Pairs,
	baseBalance float64) (
	TransactionValue float64) { // Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	// Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	TransactionValue = pair.GetLimitOnTransaction() * pair.GetLimitOnPosition() * baseBalance
	return
}

func GetBuyAndSellQuantity(
	pair pairs_interfaces.Pairs,
	baseBalance float64,
	targetBalance float64,
	buyCommission float64,
	sellCommission float64,
	ask float64,
	bid float64) (
	sellQuantity float64, // Кількість торгової валюти для продажу
	buyQuantity float64, // Кількість торгової валюти для купівлі
	err error) {
	// Кількість торгової валюти для продажу
	sellQuantity = GetTransactionValue(pair, baseBalance) / bid
	// Якщо кількість торгової валюти для продажу більша за доступну, то продаємо доступну
	if sellQuantity > targetBalance {
		sellQuantity = targetBalance
	}

	// Кількість торгової валюти для купівлі
	buyQuantity = GetTransactionValue(pair, baseBalance) / ask
	// Якщо закуплено торгової валюти більше за ліміт на позицію, то не купуємо
	if targetBalance > pair.GetLimitInputIntoPosition() {
		buyQuantity = 0
	}
	return
}

func EvaluateMiddlePrice(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs) (middlePrice float64, err error) {
	baseBalance, err := GetBaseBalance(account, pair)
	middlePrice = (pair.GetInitialBalance() - baseBalance) / (pair.GetBuyQuantity() - pair.GetSellQuantity())
	return
}

// Ініциалізація Pair
func PairInit(
	client *binance.Client,
	config *config_types.ConfigFile,
	account *spot_account.Account,
	pair pairs_interfaces.Pairs) (err error) {
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
	return nil
}

// Запускаємо потік для сбереження конфігурації кожні updateTime
func RunConfigSaver(config *config_types.ConfigFile, stopEvent chan os.Signal, updateTime time.Duration) {
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-time.After(updateTime):
				config.Save() // Зберігаємо конфігурацію кожні updateTime
			}
		}
	}()
}

func BuySellSignalInitialization(
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration) (
	buyEvent chan *depth_types.DepthItemType,
	sellEvent chan *depth_types.DepthItemType) {
	depth := depth_types.NewDepth(degree, pair.GetPair())

	bookTicker := bookTicker_types.New(degree)

	// Запускаємо потік для отримання оновлення bookTickers
	bookTickerStream := spot_streams.NewBookTickerStream(pair.GetPair(), 1)
	bookTickerStream.Start()

	triggerEvent := spot_handlers.GetBookTickersUpdateGuard(bookTicker, bookTickerStream.DataChannel)

	if updateTime > 0 {
		// Запускаємо потік для отримання оновлення BookTicker через REST
		RestBookTickerUpdater(client, stopEvent, pair, limit, updateTime, bookTicker)
		// Запускаємо потік для отримання оновлення Depth через REST
		RestDepthUpdater(client, stopEvent, pair, limit, updateTime, depth)
	}

	// Запускаємо потік для отримання сигналів на купівлю та продаж
	buyEvent, sellEvent = BuyOrSellSignal(account, depth, pair, stopEvent, triggerEvent)

	return
}

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
