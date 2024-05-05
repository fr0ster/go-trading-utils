package futures_signals

import (
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_price "github.com/fr0ster/go-trading-utils/binance/futures/markets/price"

	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	utils "github.com/fr0ster/go-trading-utils/utils"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
	price_types "github.com/fr0ster/go-trading-utils/types/price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairObserver struct {
		client             *futures.Client
		pair               pairs_interfaces.Pairs
		account            *futures_account.Account
		bookTickers        *book_ticker_types.BookTickers
		bookTickerStream   *futures_streams.BookTickerStream
		degree             int
		limit              int
		collectionOutEvent chan bool
		positionOutEvent   chan bool
		priceChanges       chan *pair_price_types.PairDelta
		riskEvent          chan bool
		priceUp            chan bool
		priceDown          chan bool
		stop               chan os.Signal
		deltaUp            float64
		deltaDown          float64
	}
)

func (pp *PairObserver) StartRiskSignal() chan bool {
	if pp.riskEvent == nil {
		pp.riskEvent = make(chan bool, 1)
		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-pp.riskEvent: // Чекаємо на спрацювання тригера
					riskPosition, err := pp.account.GetPositionRisk(pp.pair.GetPair())
					if err != nil {
						logrus.Errorf("Can't get position risk: %v, futures strategy", err)
						pp.stop <- os.Interrupt
						return
					}
					if len(riskPosition) != 1 {
						logrus.Errorf("Can't get correct position risk: %v, spot strategy", riskPosition)
						pp.stop <- os.Interrupt
						return
					}
					if (utils.ConvStrToFloat64(riskPosition[0].MarkPrice) -
						utils.ConvStrToFloat64(riskPosition[0].LiquidationPrice)/
							utils.ConvStrToFloat64(riskPosition[0].MarkPrice)) < 0.1 {
						logrus.Errorf("Risk position is too high: %v", riskPosition)
						pp.stop <- os.Interrupt
						return
					}
					pp.riskEvent <- true
				}
			}
		}()
	}
	return pp.riskEvent
}

// Запускаємо потік для оновлення ціни кожні updateTime
func (pp *PairObserver) StartPriceChangesSignal() (chan *pair_price_types.PairDelta, chan bool, chan bool) {
	if pp.priceChanges == nil && pp.priceUp == nil && pp.priceDown == nil {
		pp.priceChanges = make(chan *pair_price_types.PairDelta, 1)
		pp.priceUp = make(chan bool, 1)
		pp.priceDown = make(chan bool, 1)
		go func() {
			var (
				last_price float64
				price      *price_types.PriceChangeStats
			)
			price = price_types.New(degree)
			futures_price.Init(price, pp.client, pp.pair.GetPair())
			if priceVal := price.Get(&futures_price.SymbolPrice{Symbol: pp.pair.GetPair()}); priceVal != nil {
				last_price = utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price)
				logrus.Debugf("Start price for %s - %f", pp.pair.GetPair(), last_price)
			}
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-time.After(1 * time.Minute):
					price = price_types.New(degree)
					futures_price.Init(price, pp.client, pp.pair.GetPair())
					if priceVal := price.Get(&futures_price.SymbolPrice{Symbol: pp.pair.GetPair()}); priceVal != nil {
						if utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price) != 0 {
							current_price := utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price)
							delta := (current_price - last_price) * 100 / last_price
							logrus.Debugf("Futures, Current price for %s - %f, delta - %f", pp.pair.GetPair(), current_price, delta)
							if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
								logrus.Debugf("Futures, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
								pp.priceChanges <- &pair_price_types.PairDelta{
									Price:   utils.ConvStrToFloat64(priceVal.(*futures_price.SymbolPrice).Price),
									Percent: utils.RoundToDecimalPlace(delta, 3)}
								if delta > 0 {
									pp.priceUp <- true
								} else {
									pp.priceDown <- true
								}
								last_price = current_price
							}
						}
					}
				}
			}
		}()
	}
	return pp.priceChanges, pp.priceUp, pp.priceDown
}

func (pp *PairObserver) StartWorkInPositionSignal(triggerEvent chan bool) (collectionOutEvent chan bool) { // Виходимо з накопичення
	if pp.pair.GetStage() != pair_types.InputIntoPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.InputIntoPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if pp.collectionOutEvent == nil {
		pp.collectionOutEvent = make(chan bool, 1)

		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-triggerEvent: // Чекаємо на спрацювання тригера
				case <-time.After(pp.pair.GetTakingPositionSleepingTime()): // Або просто чекаємо якийсь час
				}
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Ліміт на вхід в позицію, відсоток від балансу базової валюти
				LimitInputIntoPosition := pp.pair.GetLimitInputIntoPosition()
				// Ліміт на позицію, відсоток від балансу базової валюти
				LimitOnPosition := pp.pair.GetLimitOnPosition()
				// Верхня межа ціни купівлі
				boundAsk,
					// Нижня межа ціни продажу
					_, err := GetBound(pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість купівлі цільової валюти більша
				// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
				// - переходимо в режим спекуляції
				if targetBalance*boundAsk >= baseBalance*LimitInputIntoPosition ||
					targetBalance*boundAsk >= baseBalance*LimitOnPosition {
					pp.pair.SetStage(pair_types.WorkInPositionStage)
					collectionOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.collectionOutEvent
}

func (pp *PairObserver) StopWorkInPositionSignal(triggerEvent chan bool) (positionOutEvent chan bool) { // Виходимо з спекуляції
	if pp.pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.WorkInPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if positionOutEvent == nil {
		pp.positionOutEvent = make(chan bool, 1)

		go func() {
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-triggerEvent: // Чекаємо на спрацювання тригера
				case <-time.After(pp.pair.GetSleepingTime()): // Або просто чекаємо якийсь час
				}
				// Кількість базової валюти
				baseBalance, err := GetBaseBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Ліміт на вхід в позицію, відсоток від балансу базової валюти
				LimitInputIntoPosition := pp.pair.GetLimitInputIntoPosition()
				// Ліміт на позицію, відсоток від балансу базової валюти
				LimitOnPosition := pp.pair.GetLimitOnPosition()
				// Верхня межа ціни купівлі
				_,
					// Нижня межа ціни продажу
					boundBid, err := GetBound(pp.pair)
				if err != nil {
					logrus.Warnf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
				if targetBalance*boundBid >= baseBalance*LimitInputIntoPosition ||
					targetBalance*boundBid >= baseBalance*LimitOnPosition {
					pp.pair.SetStage(pair_types.OutputOfPositionStage)
					positionOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.positionOutEvent
}

func NewPairObserver(
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) (pp *PairObserver, err error) {
	pp = &PairObserver{
		client:           client,
		pair:             pair,
		account:          nil,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		bookTickers:      nil,
		bookTickerStream: nil,
		priceChanges:     nil,
		priceUp:          nil,
		priceDown:        nil,
	}
	pp.account, err = futures_account.New(
		pp.client,
		pp.degree,
		[]string{pair.GetBaseSymbol()},
		[]string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
