package futures_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
	"github.com/fr0ster/go-trading-utils/utils"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
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

func RiskSignal(
	account *futures_account.Account,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) {
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				riskPosition, err := account.GetPositionRisk(pair.GetPair())
				if err != nil {
					logrus.Errorf("Can't get position risk: %v", err)
					stopEvent <- os.Interrupt
					return
				}
				if len(riskPosition) != 1 {
					logrus.Errorf("Can't get correct position risk: %v", riskPosition)
					stopEvent <- os.Interrupt
					return
				}
				if (utils.ConvStrToFloat64(riskPosition[0].MarkPrice) -
					utils.ConvStrToFloat64(riskPosition[0].LiquidationPrice)/
						utils.ConvStrToFloat64(riskPosition[0].MarkPrice)) < 0.1 {
					logrus.Errorf("Risk position is too high: %v", riskPosition)
					stopEvent <- os.Interrupt
					return
				}
				triggerEvent <- true
			}
		}
	}()
}

func PriceSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (
	increaseEvent chan *pair_price_types.PairPrice,
	decreaseEvent chan *pair_price_types.PairPrice) {
	increaseEvent = make(chan *pair_price_types.PairPrice, 1)
	decreaseEvent = make(chan *pair_price_types.PairPrice, 1)
	go func() {
		ask,
			// Ціна продажу
			bid, err := GetAskAndBid(depths)
		if err != nil {
			logrus.Errorf("Can't get data for analysis: %v", err)
			return
		}
		lastPrice := (ask + bid) / 2
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				ask,
					// Ціна продажу
					bid, err := GetAskAndBid(depths)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				currentPrice := (ask + bid) / 2
				if currentPrice > lastPrice {
					increaseEvent <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
				} else {
					decreaseEvent <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
				}
			}
			triggerEvent <- true
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
	triggerEvent chan *pair_price_types.PairPrice) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
	if pair.GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.InputIntoPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	collectionOutEvent = make(chan bool, 1)

	go func() {
		for {
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

func StopWorkingInPositionSignal(
	account *futures_account.Account,
	depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan *pair_price_types.PairPrice) (
	positionOutEvent chan bool) { // Виходимо з накопичення)
	if pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.WorkInPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	positionOutEvent = make(chan bool, 1)

	go func() {
		for {
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
