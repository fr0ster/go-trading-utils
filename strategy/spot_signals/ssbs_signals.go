package spot_signals

import (
	"context"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	account_interfaces "github.com/fr0ster/go-trading-utils/interfaces/account"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"

	pair_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

type (
	TokenInfo struct {
		CurrentProfit   float64
		PredictedProfit float64
		MiddlePrice     float64
		AvailableUSDT   float64
		Ask             float64
		Bid             float64
		BoundAsk        float64
		BoundBid        float64
	}
)

func Spot_depth_buy_sell_signals(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(account, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(account, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Ціна купівлі
				ask,
					// Ціна продажу
					bid, err := GetAskAndBid(depths)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Верхня межа ціни купівлі
				boundAsk,
					// Нижня межа ціни продажу
					boundBid, err := GetBound(pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти для продажу
				sellQuantity,
					// Кількість торгової валюти для купівлі
					buyQuantity, err := GetBuyAndSellQuantity(account, depths, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				if ((*pair).GetMiddlePrice() == 0 || // Якшо середня ціна купівли котирувальної валюти дорівнює нулю
					(*pair).GetMiddlePrice() >= boundAsk) && // Та середня ціна купівли котирувальної валюти більша або дорівнює верхній межі ціни купівли
					buyQuantity > 0 { // Та кількість цільової валюти для купівлі більша за нуль
					logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    boundAsk,
						Quantity: buyQuantity}
				} else if (*pair).GetMiddlePrice() <= boundBid && // Якшо середня ціна купівли котирувальної валюти менша або дорівнює нижній межі ціни продажу
					sellQuantity > 0 { // Та кількість цільової валюти для продажу більша за нуль
					logrus.Infof("Middle price %f is lower than low bound price %f, SELL!!!", (*pair).GetMiddlePrice(), boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
				}
				logrus.Infof("Now ask is %f, bid is %f", ask, bid)
				logrus.Infof("Current Ask bound: %f, Bid bound: %f", boundAsk, boundBid)
				logrus.Infof("Middle price: %f, available USDT: %f, available %s: %f",
					(*pair).GetMiddlePrice(), baseBalance, (*pair).GetTargetSymbol(), targetBalance)
				logrus.Infof("Current profit: %f", (*pair).GetProfit(bid))
				logrus.Infof("Predicable profit: %f", (*pair).GetProfit((*pair).GetMiddlePrice()*(1+(*pair).GetSellDelta())))
				time.Sleep(5 * time.Second)
			}
		}
	}()
	return
}

func BuyOrSellSignal(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	stopByOrSell chan bool,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	if (*pair).GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", (*pair).GetStage(), pair_types.WorkInPositionStage)
		return
	}
	go func() {
		for {
			if (*pair).GetMiddlePrice() == 0 {
				continue
			}
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера

				ask,
					// Ціна продажу
					bid, err := GetAskAndBid(depths)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Верхня межа ціни купівлі
				boundAsk,
					// Нижня межа ціни продажу
					boundBid, err := GetBound(pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти для продажу
				sellQuantity,
					// Кількість торгової валюти для купівлі
					buyQuantity, err := GetBuyAndSellQuantity(account, depths, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Середня ціна купівли котирувальної валюти дорівнює нулю або більша за верхню межу ціни купівли
				if (*pair).GetMiddlePrice() >= boundAsk &&
					buyQuantity > 0 { // Та кількість цільової валюти для купівлі більша за нуль
					logrus.Infof("Middle price %f is higher than high bound price %f, BUY!!!", (*pair).GetMiddlePrice(), boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    boundAsk,
						Quantity: buyQuantity}
					// Середня ціна купівли котирувальної валюти менша або дорівнює нижній межі ціни продажу
				} else if (*pair).GetMiddlePrice() <= boundBid &&
					sellQuantity > 0 { // Та кількість цільової валюти для продажу більша за нуль
					logrus.Infof("Middle price %f is lower than low bound price %f, SELL!!!", (*pair).GetMiddlePrice(), boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
				} else {
					// Чекаємо на зміну ціни
					logrus.Infof("Middle price is %f, bound Bid price %f, bound Ask price %f",
						(*pair).GetMiddlePrice(), boundBid, boundAsk)
					if buyQuantity == 0 || sellQuantity == 0 {
						logrus.Info("Wait for buy signal")
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for ask decrease to %f", boundAsk)
					} else if (*pair).GetMiddlePrice() > boundBid && (*pair).GetMiddlePrice() < boundAsk {
						logrus.Info("Wait for buy or sell signal")
						logrus.Infof("Now ask is %f, bid is %f", ask, bid)
						logrus.Infof("Waiting for ask decrease to %f or bid increase to %f", boundAsk, boundBid)
					}
				}
			}
		}
	}()
	return
}

func InitMiddlePrice(
	client *binance.Client,
	pair *config_interfaces.Pairs,
	pairInfo *symbol_info_types.Symbol) (err error) {
	quantity := (*pairInfo).LotSizeFilter().MinQuantity

	if (*pair).GetMiddlePrice() == 0 {
		_, err =
			client.NewCreateOrderService().
				Symbol(string(binance.SymbolType((*pair).GetPair()))).
				Type(binance.OrderTypeMarket).
				Side(binance.SideTypeBuy).
				Quantity(quantity).
				TimeInForce(binance.TimeInForceTypeGTC).Do(context.Background())
	}
	return
}

func StartWorkInPositionSignal(
	client *binance.Client,
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan *depth_types.DepthItemType) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
	if (*pair).GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", (*pair).GetStage(), pair_types.InputIntoPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	collectionOutEvent = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
			case <-time.After((*pair).GetSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Кількість базової валюти
			baseBalance, err := GetBaseBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Ліміт на вхід в позицію, відсоток від балансу базової валюти
			LimitInputIntoPosition := (*pair).GetLimitInputIntoPosition()
			// Ліміт на позицію, відсоток від балансу базової валюти
			LimitOnPosition := (*pair).GetLimitOnPosition()
			// Верхня межа ціни купівлі
			boundAsk,
				// Нижня межа ціни продажу
				_, err := GetBound(pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість купівлі цільової валюти більша
			// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
			// - переходимо в режим спекуляції
			if targetBalance*boundAsk >= baseBalance*LimitInputIntoPosition ||
				targetBalance*boundAsk >= baseBalance*LimitOnPosition {
				(*pair).SetStage(pair_types.WorkInPositionStage)
				collectionOutEvent <- true
				return
			}
		}
	}()
	return
}

func StartOutputOfPositionSignal(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	buyEvent chan *depth_types.DepthItemType) (
	positionOutEvent chan bool) { // Виходимо з накопичення)
	if (*pair).GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", (*pair).GetStage(), pair_types.WorkInPositionStage)
		return
	}

	positionOutEvent = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-buyEvent: // Чекаємо на спрацювання тригера
			case <-time.After((*pair).GetSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Кількість базової валюти
			baseBalance, err := GetBaseBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Ліміт на вхід в позицію, відсоток від балансу базової валюти
			LimitInputIntoPosition := (*pair).GetLimitInputIntoPosition()
			// Ліміт на позицію, відсоток від балансу базової валюти
			LimitOnPosition := (*pair).GetLimitOnPosition()
			// Верхня межа ціни купівлі
			_,
				// Нижня межа ціни продажу
				boundBid, err := GetBound(pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
			if targetBalance*boundBid >= baseBalance*LimitInputIntoPosition ||
				targetBalance*boundBid >= baseBalance*LimitOnPosition {
				(*pair).SetStage(pair_types.OutputOfPositionStage)
				positionOutEvent <- true
				return
			}
		}
	}()
	return
}

func StopWorkingSignal(
	account account_interfaces.Accounts,
	depths *depth_types.Depth,
	pair *config_interfaces.Pairs,
	stopEvent chan os.Signal,
	buyEvent chan *depth_types.DepthItemType) (
	stopWorkingEvent chan bool) { // Виходимо з накопичення)
	if (*pair).GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", (*pair).GetStage(), pair_types.WorkInPositionStage)
		return
	}

	stopWorkingEvent = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-stopEvent:
				return
			case <-buyEvent: // Чекаємо на спрацювання тригера
			case <-time.After((*pair).GetSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
			if targetBalance == 0 {
				(*pair).SetStage(pair_types.WorkInPositionStage)
				stopWorkingEvent <- true
				return
			}
		}
	}()
	return
}
