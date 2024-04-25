package spot_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	book_types "github.com/fr0ster/go-trading-utils/types/bookticker"
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

func PriceSignal(
	bookTickers *book_types.BookTickers,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (
	increaseEvent chan *pair_price_types.PairPrice,
	decreaseEvent chan *pair_price_types.PairPrice,
	waitingEvent chan *pair_price_types.PairPrice) {
	var lastPrice float64
	increaseEvent = make(chan *pair_price_types.PairPrice, 1)
	decreaseEvent = make(chan *pair_price_types.PairPrice, 1)
	waitingEvent = make(chan *pair_price_types.PairPrice, 1)
	bookTicker := bookTickers.Get(pair.GetPair())
	if bookTicker == nil {
		logrus.Errorf("Can't get bookTicker for %s when read for last price, spot strategy", pair.GetPair())
		stopEvent <- os.Interrupt
		return
	}
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-triggerEvent: // Чекаємо на спрацювання тригера
				bookTicker := bookTickers.Get(pair.GetPair())
				if bookTicker == nil {
					logrus.Errorf("Can't get bookTicker for %s", pair.GetPair())
					stopEvent <- os.Interrupt
					return
				}
				// Ціна купівлі
				ask := bookTicker.(*book_types.BookTicker).AskPrice
				// Ціна продажу
				bid := bookTicker.(*book_types.BookTicker).AskPrice
				if lastPrice == 0 {
					lastPrice = (ask + bid) / 2
					continue
				}
				currentPrice := (ask + bid) / 2
				if currentPrice > lastPrice {
					increaseEvent <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
					lastPrice = currentPrice
				} else if currentPrice < lastPrice {
					decreaseEvent <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
					lastPrice = currentPrice
				} else {
					waitingEvent <- &pair_price_types.PairPrice{
						Price: currentPrice,
					}
				}
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func BuyOrSellSignal(
	account *spot_account.Account,
	bookTickers *book_types.BookTickers,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	triggerEvent chan bool) (
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) {
	buyEvent = make(chan *pair_price_types.PairPrice, 1)
	sellEvent = make(chan *pair_price_types.PairPrice, 1)
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
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(account, pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				pair.SetCurrentBalance(baseBalance)
				pair.SetCurrentPositionBalance(baseBalance * pair.GetLimitOnPosition())
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(account, pair)
				if err != nil {
					logrus.Errorf("Can't get %s balance: %v", pair.GetTargetSymbol(), err)
					stopEvent <- os.Interrupt
					return
				}
				commission := GetCommission(account)
				bookTicker := bookTickers.Get(pair.GetPair())
				if bookTicker == nil {
					logrus.Errorf("Can't get bookTicker for %s", pair.GetPair())
					stopEvent <- os.Interrupt
					return
				}
				// Ціна купівлі
				ask := bookTicker.(*book_types.BookTicker).AskPrice
				// Ціна продажу
				bid := bookTicker.(*book_types.BookTicker).AskPrice
				// Верхня межа ціни купівлі
				boundAsk, err := GetAskBound(pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					stopEvent <- os.Interrupt
					return
				}
				// Нижня межа ціни продажу
				boundBid, err := GetBidBound(pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					stopEvent <- os.Interrupt
					return
				}
				// Кількість торгової валюти для продажу
				sellQuantity,
					// Кількість торгової валюти для купівлі
					buyQuantity, err := GetBuyAndSellQuantity(pair, baseBalance, targetBalance, commission, commission, ask, bid)
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
					logrus.Debugf("Middle price %f, Ask %f is lower than high bound price %f, BUY!!!", pair.GetMiddlePrice(), ask, boundAsk)
					buyEvent <- &pair_price_types.PairPrice{
						Price:    ask,
						Quantity: buyQuantity}
					// Середня ціна купівли цільової валюти менша або дорівнює нижній межі ціни продажу
				} else if bid >= boundBid && sellQuantity < targetBalance {
					logrus.Debugf("Middle price %f, Bid %f is higher than low bound price %f, SELL!!!", pair.GetMiddlePrice(), bid, boundBid)
					sellEvent <- &pair_price_types.PairPrice{
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
	account *spot_account.Account,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) (
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
			case <-buyEvent: // Чекаємо на спрацювання тригера на купівлю
			case <-sellEvent: // Чекаємо на спрацювання тригера на продаж
			case <-time.After(pair.GetTakingPositionSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Кількість базової валюти
			baseBalance, err := GetBaseBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			pair.SetCurrentBalance(baseBalance)
			pair.SetCurrentPositionBalance(baseBalance * pair.GetLimitOnPosition())
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Errorf("Can't get data for analysis: %v", err)
				continue
			}
			// Верхня межа ціни купівлі
			boundAsk, err := GetAskBound(pair)
			if err != nil {
				logrus.Errorf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість купівлі цільової валюти більша
			// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
			// - переходимо в режим спекуляції
			if targetBalance*boundAsk >= baseBalance*pair.GetLimitInputIntoPosition() ||
				targetBalance*boundAsk >= baseBalance*pair.GetLimitOnPosition() {
				pair.SetStage(pair_types.WorkInPositionStage)
				collectionOutEvent <- true
				return
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}

func StopWorkInPositionSignal(
	account *spot_account.Account,
	// depths *depth_types.Depth,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	buyEvent chan *pair_price_types.PairPrice,
	sellEvent chan *pair_price_types.PairPrice) (
	workingOutEvent chan bool) { // Виходимо з накопичення
	if pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pair.GetStage(), pair_types.WorkInPositionStage)
		stopEvent <- os.Interrupt
		return
	}

	workingOutEvent = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent: // Чекаємо на спрацювання тригера на купівлю
			case <-sellEvent: // Чекаємо на спрацювання тригера на продаж
			case <-time.After(pair.GetSleepingTime()): // Або просто чекаємо якийсь час
			}
			// Ліміт на вхід в позицію, відсоток від балансу базової валюти
			// LimitInputIntoPosition := pair.GetLimitInputIntoPosition()
			// Ліміт на позицію, відсоток від балансу базової валюти
			// LimitOnPosition := pair.GetLimitOnPosition()
			// Кількість базової валюти
			baseBalance, err := GetBaseBalance(account, pair)
			if err != nil {
				logrus.Warnf("Can't get data for analysis: %v", err)
				continue
			}
			pair.SetCurrentBalance(baseBalance)
			pair.SetCurrentPositionBalance(baseBalance * pair.GetLimitOnPosition())
			// Кількість торгової валюти
			targetBalance, err := GetTargetBalance(account, pair)
			if err != nil {
				logrus.Errorf("Can't get data for analysis: %v", err)
				continue
			}
			// Нижня межа ціни продажу
			boundBid, err := GetBidBound(pair)
			if err != nil {
				logrus.Errorf("Can't get data for analysis: %v", err)
				continue
			}
			// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
			if targetBalance*boundBid >= baseBalance*pair.GetLimitInputIntoPosition() ||
				targetBalance*boundBid >= baseBalance*pair.GetLimitOnPosition() {
				pair.SetStage(pair_types.OutputOfPositionStage)
				workingOutEvent <- true
				return
			}
			time.Sleep(pair.GetSleepingTime())
		}
	}()
	return
}
