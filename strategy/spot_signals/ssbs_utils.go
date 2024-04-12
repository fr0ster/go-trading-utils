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

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/info"
	spot_bookticker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	bookTicker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
)

const (
	errorMsg = "Error: %v"
)

func LimitRead(degree int, symbols []string, client *binance.Client) (
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits) {
	exchangeInfo := exchange_types.NewExchangeInfo()
	spot_exchange_info.RestrictedInit(exchangeInfo, degree, symbols, client)

	minuteOrderLimit = exchangeInfo.Get_Minute_Order_Limit()
	dayOrderLimit = exchangeInfo.Get_Day_Order_Limit()
	minuteRawRequestLimit = exchangeInfo.Get_Minute_Raw_Request_Limit()
	updateTime = minuteRawRequestLimit.Interval * time.Duration(1+minuteRawRequestLimit.IntervalNum)
	return
}

func RestUpdate(
	client *binance.Client,
	stop chan os.Signal,
	pair *config_interfaces.Pairs,
	depth *depth_types.Depth,
	limit int,
	bookTicker *bookTicker_types.BookTickerBTree,
	updateTime time.Duration) {
	go func() {
		for {
			select {
			case <-stop:
				// Якщо отримано сигнал з каналу stop, вийти з циклу
				return
			default:
				err := spot_depth.SpotDepthInit(depth, client, limit)
				if err != nil {
					logrus.Errorf(errorMsg, err)
					stop <- os.Interrupt
					return
				}

				time.Sleep(1 * time.Second)

				err = spot_bookticker.Init(bookTicker, (*pair).GetPair(), client)
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

func GetBound(pair *config_interfaces.Pairs) (boundAsk float64, boundBid float64, err error) {
	boundAsk = (*pair).GetMiddlePrice() * (1 - (*pair).GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	boundBid = (*pair).GetMiddlePrice() * (1 + (*pair).GetSellDelta())
	logrus.Debugf("Bid bound: %f", boundBid)
	return
}

func GetBaseBalance(
	account *account_interfaces.Accounts,
	pair *config_interfaces.Pairs) (
	baseBalance float64, // Кількість базової валюти
	err error) {
	baseBalance, err = func(pair *config_interfaces.Pairs) (
		baseBalance float64,
		err error) {
		baseBalance, err = (*account).GetAsset((*pair).GetBaseSymbol())
		return
	}(pair)

	if err != nil {
		return 0, err
	}
	return
}

func GetTargetBalance(
	account *account_interfaces.Accounts,
	pair *config_interfaces.Pairs) (
	targetBalance float64, // Кількість торгової валюти
	err error) {
	targetBalance, err = func(pair *config_interfaces.Pairs) (
		targetBalance float64,
		err error) {
		targetBalance, err = (*account).GetAsset((*pair).GetTargetSymbol())
		return
	}(pair)

	if err != nil {
		return 0, err
	}
	return
}

func GetTransactionValue(
	pair *config_interfaces.Pairs,
	baseBalance float64) (
	TransactionValue float64) { // Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	// Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	TransactionValue = (*pair).GetLimitOnTransaction() * (*pair).GetLimitOnPosition() * baseBalance
	return
}

func GetBuyAndSellQuantity(
	account *account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	baseBalance float64,
	targetBalance float64) (
	sellQuantity float64, // Кількість торгової валюти для продажу
	buyQuantity float64, // Кількість торгової валюти для купівлі
	err error) {
	bid, _, err := GetAskAndBid(depths)
	if err != nil {
		return
	}
	boundAsk, _, err := GetBound(pair)
	if err != nil {
		return
	}
	// Кількість торгової валюти для продажу
	sellQuantity = GetTransactionValue(pair, baseBalance) / bid
	// Якщо кількість торгової валюти для продажу більша за доступну, то продаємо доступну
	if sellQuantity > targetBalance {
		sellQuantity = targetBalance
	}

	// Кількість торгової валюти для купівлі
	buyQuantity = GetTransactionValue(pair, baseBalance) / boundAsk
	// Якщо закуплено торгової валюти більше за ліміт на позицію, то не купуємо
	if targetBalance > (*pair).GetLimitInputIntoPosition() {
		buyQuantity = 0
	}
	return
}

func EvaluateMiddlePrice(
	account *account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs) (middlePrice float64, err error) {
	err = (*account).Update()
	if err != nil {
		return
	}
	baseBalance, err := GetBaseBalance(account, pair)
	middlePrice = ((*pair).GetInitialBalance() - baseBalance) / ((*pair).GetBuyQuantity() - (*pair).GetSellQuantity())
	return
}
