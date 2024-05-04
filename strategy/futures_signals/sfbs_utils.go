package futures_signals

import (
	"context"
	"errors"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_bookticker "github.com/fr0ster/go-trading-utils/binance/futures/markets/bookticker"
	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
)

func LimitRead(degree int, symbols []string, client *futures.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	exchangeInfo := exchange_types.New()
	futures_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}

func RestBookTickerUpdater(
	client *futures.Client,
	stop chan os.Signal,
	pair pairs_interfaces.Pairs,
	limit int,
	updateTime time.Duration,
	bookTicker *book_ticker_types.BookTickers) {
	go func() {
		for {
			select {
			case <-stop:
				// Якщо отримано сигнал з каналу stop, вийти з циклу
				return
			default:
				err := futures_bookticker.Init(bookTicker, pair.GetPair(), client)
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
	client *futures.Client,
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
				err := futures_depth.Init(depth, client, limit)
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

func GetPrice(client *futures.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func getPrice(val btree.Item) (float64, error) {
	if val == nil {
		return 0, errors.New("value is nil")
	}
	return val.(*pair_price_types.PairPrice).Price, nil
}

func GetCommission(
	account account_interfaces.Accounts) (
	commission float64) { // Комісія за покупку/Комісія за продаж)
	return account.GetMakerCommission()
}

func GetDepthAsk(depths *depth_types.Depth) (ask float64, err error) {
	ask, err = getPrice(depths.GetAsks().Min())
	return
}

func GetDepthBid(depths *depth_types.Depth) (bid float64, err error) {
	bid, err = getPrice(depths.GetBids().Max())
	return
}

func GetBookTickerAsk(bookTicker *book_ticker_types.BookTicker) (price float64, quantity float64) {
	return bookTicker.AskPrice, bookTicker.AskQuantity
}

func GetBookTickerBid(bookTicker *book_ticker_types.BookTicker) (price float64, quantity float64) {
	return bookTicker.BidPrice, bookTicker.BidQuantity
}

func GetBound(pair pairs_interfaces.Pairs) (boundAsk float64, boundBid float64, err error) {
	boundAsk = pair.GetMiddlePrice() * (1 - pair.GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	boundBid = pair.GetMiddlePrice() * (1 + pair.GetSellDelta())
	logrus.Debugf("Bid bound: %f", boundBid)
	return
}

func GetAskBound(pair pairs_interfaces.Pairs) (boundAsk float64, err error) {
	boundAsk = pair.GetMiddlePrice() * (1 - pair.GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	return
}

func GetBidBound(pair pairs_interfaces.Pairs) (boundBid float64, err error) {
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
	account *futures_account.Account,
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
	client *futures.Client,
	config *config_types.ConfigFile,
	account *futures_account.Account,
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
