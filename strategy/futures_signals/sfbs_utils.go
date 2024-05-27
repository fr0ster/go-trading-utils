package futures_signals

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
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

func GetPrice(client *futures.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func GetCommission(
	account account_interfaces.Accounts) (
	commission float64) { // Комісія за покупку/Комісія за продаж)
	return account.GetMakerCommission()
}

func GetBound(pair *pairs_types.Pairs) (boundAsk float64, boundBid float64, err error) {
	boundAsk = pair.GetMiddlePrice() * (1 - pair.GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	boundBid = pair.GetMiddlePrice() * (1 + pair.GetSellDelta())
	logrus.Debugf("Bid bound: %f", boundBid)
	return
}

func GetAskBound(pair *pairs_types.Pairs) (boundAsk float64, err error) {
	boundAsk = pair.GetMiddlePrice() * (1 - pair.GetBuyDelta())
	logrus.Debugf("Ask bound: %f", boundAsk)
	return
}

func GetBidBound(pair *pairs_types.Pairs) (boundBid float64, err error) {
	boundBid = pair.GetMiddlePrice() * (1 + pair.GetSellDelta())
	logrus.Debugf("Bid bound: %f", boundBid)
	return
}

func GetBaseBalance(
	account account_interfaces.Accounts,
	pair *pairs_types.Pairs) (
	baseBalance float64, // Кількість базової валюти
	err error) {
	baseBalance, err = func(pair *pairs_types.Pairs) (
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
	pair *pairs_types.Pairs) (
	targetBalance float64, // Кількість торгової валюти
	err error) {
	targetBalance, err = func(pair *pairs_types.Pairs) (
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
	pair *pairs_types.Pairs,
	baseBalance float64) (
	TransactionValue float64) { // Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	// Сума для транзакції, множимо баланс базової валюти на ліміт на транзакцію та на ліміт на позицію
	TransactionValue = pair.GetLimitOnTransaction() * pair.GetLimitOnPosition() * baseBalance
	return
}

func GetBuyAndSellQuantity(
	pair *pairs_types.Pairs,
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
