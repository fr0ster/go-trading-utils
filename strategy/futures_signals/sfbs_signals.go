package futures_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
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

func BuyOrSellSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (buyEvent chan *depth_types.DepthItemType, sellEvent chan *depth_types.DepthItemType) {
	buyEvent = make(chan *depth_types.DepthItemType, 1)
	sellEvent = make(chan *depth_types.DepthItemType, 1)
	go func() {
		for {
			if pair.GetMiddlePrice() == 0 {
				continue
			}
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				err := account.Update()
				if err != nil {
					logrus.Errorf("Can't update account: %v", err)
					stopEvent <- os.Interrupt
					return
				}
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(account, pair)
				if err != nil {
					logrus.Errorf("Can't get %s balance: %v", pair.GetTargetSymbol(), err)
					stopEvent <- os.Interrupt
					return
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(account, pair)
				if err != nil {
					logrus.Errorf("Can't get %s balance: %v", pair.GetTargetSymbol(), err)
					stopEvent <- os.Interrupt
					return
				}
				ask,
					// Ціна продажу
					bid, err := GetAskAndBid(depths)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Верхня межа ціни купівлі
				boundAsk,
					// Нижня межа ціни продажу
					boundBid, err := GetBound(pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					stopEvent <- os.Interrupt
					return
				}
				// Кількість торгової валюти для продажу
				sellQuantity,
					// Кількість торгової валюти для купівлі
					buyQuantity, err := GetBuyAndSellQuantity(pair, baseBalance, targetBalance, ask, bid)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					stopEvent <- os.Interrupt
					return
				}

				if buyQuantity == 0 && sellQuantity == 0 {
					logrus.Errorf("We don't have any %s for buy and don't have any %s for sell",
						pair.GetBaseSymbol(), pair.GetTargetSymbol())
					stopEvent <- os.Interrupt
					return
				}
				// Середня ціна купівли цільової валюти більша за верхню межу ціни купівли
				if ask <= boundAsk &&
					targetBalance*ask < pair.GetLimitInputIntoPosition()*baseBalance &&
					targetBalance*ask < pair.GetLimitOutputOfPosition()*baseBalance {
					logrus.Infof("Middle price %f, Ask %f is lower than high bound price %f, BUY!!!", pair.GetMiddlePrice(), ask, boundAsk)
					buyEvent <- &depth_types.DepthItemType{
						Price:    ask,
						Quantity: buyQuantity}
					// Середня ціна купівли цільової валюти менша або дорівнює нижній межі ціни продажу
				} else if bid >= boundBid && sellQuantity < targetBalance {
					logrus.Infof("Middle price %f, Bid %f is higher than low bound price %f, SELL!!!", pair.GetMiddlePrice(), bid, boundBid)
					sellEvent <- &depth_types.DepthItemType{
						Price:    boundBid,
						Quantity: sellQuantity}
				} else {
					if ask <= boundAsk &&
						(targetBalance*ask > pair.GetLimitInputIntoPosition()*baseBalance ||
							targetBalance*ask > pair.GetLimitOutputOfPosition()*baseBalance) {
						logrus.Debugf("We can't buy %s, because we have more than %f %s",
							pair.GetTargetSymbol(),
							pair.GetLimitInputIntoPosition()*baseBalance,
							pair.GetBaseSymbol())
					} else if bid >= boundBid && sellQuantity >= targetBalance {
						logrus.Debugf("We can't sell %s, because we haven't %s enough for sell, we need %f %s but have %f %s only",
							pair.GetTargetSymbol(),
							pair.GetTargetSymbol(),
							sellQuantity,
							pair.GetTargetSymbol(),
							targetBalance,
							pair.GetTargetSymbol())
					} else if bid < boundBid && ask > boundAsk { // Чекаємо на зміну ціни
						logrus.Debugf("Middle price is %f, bound Bid price %f, bound Ask price %f",
							pair.GetMiddlePrice(), boundBid, boundAsk)
						logrus.Debugf("Wait for buy or sell signal")
						logrus.Debugf("Now ask is %f, bid is %f", ask, bid)
						logrus.Debugf("Waiting for ask decrease to %f or bid increase to %f", boundAsk, boundBid)
					}
				}
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func StartWorkInPositionSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan *depth_types.DepthItemType) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
	if pair.GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.InputIntoPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	collectionOutEvent = make(chan bool, 1)

	go func() {
		for {
			err := account.Update()
			if err != nil {
				logrus.Errorf("Can't update account: %v", err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
			case <-time.After(pair.GetTakingPositionSleepingTime()): // Або просто чекаємо якийсь час
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
			LimitInputIntoPosition := pair.GetLimitInputIntoPosition()
			// Ліміт на позицію, відсоток від балансу базової валюти
			LimitOnPosition := pair.GetLimitOnPosition()
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
				pair.SetStage(pair_types.WorkInPositionStage)
				collectionOutEvent <- true
				return
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func StartOutputOfPositionSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan *depth_types.DepthItemType) (
	positionOutEvent chan bool) { // Виходимо з накопичення)
	if pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.WorkInPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	positionOutEvent = make(chan bool, 1)

	go func() {
		for {
			err := account.Update()
			if err != nil {
				logrus.Errorf("Can't update account: %v", err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
			case <-time.After(pair.GetSleepingTime()): // Або просто чекаємо якийсь час
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
			LimitInputIntoPosition := pair.GetLimitInputIntoPosition()
			// Ліміт на позицію, відсоток від балансу базової валюти
			LimitOnPosition := pair.GetLimitOnPosition()
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
				pair.SetStage(pair_types.OutputOfPositionStage)
				positionOutEvent <- true
				return
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func StopWorkingSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan *depth_types.DepthItemType) (
	stopWorkingEvent chan bool) { // Виходимо з накопичення)
	if pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.WorkInPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	stopWorkingEvent = make(chan bool, 1)

	go func() {
		for {
			err := account.Update()
			if err != nil {
				logrus.Errorf("Can't update account: %v", err)
				stopEvent <- os.Interrupt
				return
			}
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
			case <-time.After(pair.GetSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
			if targetBalance == 0 {
				pair.SetStage(pair_types.WorkInPositionStage)
				stopWorkingEvent <- true
				return
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}
