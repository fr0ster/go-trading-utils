package spot_signals

// import (
// 	_ "net/http/pprof"
// 	"time"

// 	"os"

// 	"github.com/sirupsen/logrus"

// 	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

// 	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

// 	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
// )

// type (
// 	TokenInfo struct {
// 		CurrentProfit   float64
// 		PredictedProfit float64
// 		MiddlePrice     float64
// 		AvailableUSDT   float64
// 		Ask             float64
// 		Bid             float64
// 		BoundAsk        float64
// 		BoundBid        float64
// 	}
// )

// func StartWorkInPositionSignal(
// 	account *spot_account.Account,
// 	pair pairs_interfaces.Pairs,
// 	stopEvent chan os.Signal,
// 	triggerEvent chan bool) (
// 	collectionOutEvent chan bool) { // Виходимо з накопичення
// 	if pair.GetStage() != pair_types.InputIntoPositionStage {
// 		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.InputIntoPositionStage)
// 		stopEvent <- os.Interrupt
// 		return
// 	}

// 	collectionOutEvent = make(chan bool, 1)

// 	go func() {
// 		for {
// 			select {
// 			case <-stopEvent:
// 				stopEvent <- os.Interrupt
// 				return
// 			case <-triggerEvent: // Чекаємо на спрацювання тригера
// 			case <-time.After(pair.GetTakingPositionSleepingTime()): // Або просто чекаємо якийсь час
// 			}
// 			// Кількість базової валюти
// 			baseBalance, err := GetBaseBalance(account, pair)
// 			if err != nil {
// 				logrus.Warnf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			pair.SetCurrentBalance(baseBalance)
// 			pair.SetCurrentPositionBalance(baseBalance * pair.GetLimitOnPosition())
// 			// Кількість торгової валюти
// 			targetBalance, err := GetTargetBalance(account, pair)
// 			if err != nil {
// 				logrus.Errorf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			// Верхня межа ціни купівлі
// 			boundAsk, err := GetAskBound(pair)
// 			if err != nil {
// 				logrus.Errorf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			// Якшо вартість купівлі цільової валюти більша
// 			// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
// 			// - переходимо в режим спекуляції
// 			if targetBalance*boundAsk >= baseBalance*pair.GetLimitInputIntoPosition() {
// 				pair.SetStage(pair_types.WorkInPositionStage)
// 				collectionOutEvent <- true
// 				return
// 			}
// 			time.Sleep(pair.GetSleepingTime())
// 		}
// 	}()
// 	return
// }

// func StopWorkInPositionSignal(
// 	account *spot_account.Account,
// 	// depths *depth_types.Depth,
// 	pair pairs_interfaces.Pairs,
// 	stopEvent chan os.Signal,
// 	triggerEvent chan bool) (
// 	workingOutEvent chan bool) { // Виходимо з накопичення
// 	if pair.GetStage() != pair_types.WorkInPositionStage {
// 		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.WorkInPositionStage)
// 		stopEvent <- os.Interrupt
// 		return
// 	}

// 	workingOutEvent = make(chan bool, 1)

// 	go func() {
// 		for {
// 			select {
// 			case <-stopEvent:
// 				stopEvent <- os.Interrupt
// 				return
// 			case <-triggerEvent: // Чекаємо на спрацювання тригера
// 			case <-time.After(pair.GetSleepingTime()): // Або просто чекаємо якийсь час
// 			}
// 			// Ліміт на вхід в позицію, відсоток від балансу базової валюти
// 			// LimitInputIntoPosition := pair.GetLimitInputIntoPosition()
// 			// Ліміт на позицію, відсоток від балансу базової валюти
// 			// LimitOnPosition := pair.GetLimitOnPosition()
// 			// Кількість базової валюти
// 			baseBalance, err := GetBaseBalance(account, pair)
// 			if err != nil {
// 				logrus.Warnf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			pair.SetCurrentBalance(baseBalance)
// 			pair.SetCurrentPositionBalance(baseBalance * pair.GetLimitOnPosition())
// 			// Кількість торгової валюти
// 			targetBalance, err := GetTargetBalance(account, pair)
// 			if err != nil {
// 				logrus.Errorf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			// Нижня межа ціни продажу
// 			boundBid, err := GetBidBound(pair)
// 			if err != nil {
// 				logrus.Errorf("Can't get data for analysis: %v", err)
// 				continue
// 			}
// 			// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
// 			if targetBalance*boundBid >= baseBalance*pair.GetLimitOutputOfPosition() {
// 				pair.SetStage(pair_types.OutputOfPositionStage)
// 				workingOutEvent <- true
// 				return
// 			}
// 			time.Sleep(pair.GetSleepingTime())
// 		}
// 	}()
// 	return
// }
