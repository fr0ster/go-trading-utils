package spot_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_price "github.com/fr0ster/go-trading-utils/binance/spot/markets/price"

	"github.com/fr0ster/go-trading-utils/utils"

	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	book_ticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"
	price_types "github.com/fr0ster/go-trading-utils/types/price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairObserver struct {
		client             *binance.Client
		pair               pairs_interfaces.Pairs
		degree             int
		limit              int
		account            *spot_account.Account
		bookTickers        *book_ticker_types.BookTickers
		bookTickerStream   *spot_streams.BookTickerStream
		bookTickerEvent    chan bool
		depths             *depth_types.Depth
		depthsStream       *spot_streams.DepthStream
		depthEvent         chan bool
		klines             *kline_types.Klines
		klineStream        *spot_streams.KlineStream
		klineEvent         chan bool
		priceChanges       chan *pair_price_types.PairDelta
		collectionOutEvent chan bool
		workingOutEvent    chan bool
		priceUp            chan bool
		priceDown          chan bool
		stop               chan os.Signal
		deltaUp            float64
		deltaDown          float64
	}
)

// Запускаємо потік для оновлення ціни кожні updateTime
func (pp *PairObserver) StartPriceChangesSignal() (chan *pair_price_types.PairDelta, chan bool, chan bool) {
	if pp.priceChanges == nil && pp.priceUp == nil && pp.priceDown == nil {
		pp.priceChanges = make(chan *pair_price_types.PairDelta, 1)
		pp.priceUp = make(chan bool, 1)
		pp.priceDown = make(chan bool, 1)
		go func() {
			var (
				price      *price_types.PriceChangeStats
				last_price float64
			)
			price = price_types.New(degree)
			spot_price.Init(price, pp.client, pp.pair.GetPair())
			if priceVal := price.Get(&spot_price.SymbolTicker{Symbol: pp.pair.GetPair()}); priceVal != nil {
				last_price = utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice)
				logrus.Debugf("Start price for %s - %f", pp.pair.GetPair(), last_price)
			}
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-time.After(1 * time.Minute):
					price = price_types.New(degree)
					spot_price.Init(price, pp.client, pp.pair.GetPair())
					if priceVal := price.Get(&spot_price.SymbolTicker{Symbol: pp.pair.GetPair()}); priceVal != nil {
						if utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice) != 0 {
							current_price := utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice)
							delta := func() float64 { return (current_price - last_price) * 100 / last_price }
							if last_price != 0 {
								logrus.Debugf("Spot, Current price for %s - %f, delta - %f", pp.pair.GetPair(), current_price, delta())
							}
							if delta() > pp.deltaUp*100 || delta() < -pp.deltaDown*100 {
								logrus.Debugf("Spot, Price for %s is changed on %f%%", pp.pair.GetPair(), delta())
								pp.priceChanges <- &pair_price_types.PairDelta{
									Price:   utils.ConvStrToFloat64(priceVal.(*spot_price.SymbolTicker).LastPrice),
									Percent: utils.RoundToDecimalPlace(delta(), 3)}
								if delta() > 0 {
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

func (pp *PairObserver) StartWorkInPositionSignal(triggerEvent chan bool) (
	collectionOutEvent chan bool) { // Виходимо з накопичення
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
				pp.pair.SetCurrentBalance(baseBalance)
				pp.pair.SetCurrentPositionBalance(baseBalance * pp.pair.GetLimitOnPosition())
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Верхня межа ціни купівлі
				boundAsk, err := GetAskBound(pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість купівлі цільової валюти більша
				// за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію
				// - переходимо в режим спекуляції
				if targetBalance*boundAsk >= baseBalance*pp.pair.GetLimitInputIntoPosition() {
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

func (pp *PairObserver) StopWorkInPositionSignal(triggerEvent chan bool) (
	workingOutEvent chan bool) { // Виходимо з спекуляції
	if pp.pair.GetStage() != pair_types.WorkInPositionStage {
		logrus.Errorf("Strategy stage %s is not %s", pp.pair.GetStage(), pair_types.WorkInPositionStage)
		pp.stop <- os.Interrupt
		return
	}

	if pp.workingOutEvent == nil {
		pp.workingOutEvent = make(chan bool, 1)

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
				pp.pair.SetCurrentBalance(baseBalance)
				pp.pair.SetCurrentPositionBalance(baseBalance * pp.pair.GetLimitOnPosition())
				// Кількість торгової валюти
				targetBalance, err := GetTargetBalance(pp.account, pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Нижня межа ціни продажу
				boundBid, err := GetBidBound(pp.pair)
				if err != nil {
					logrus.Errorf("Can't get data for analysis: %v", err)
					continue
				}
				// Якшо вартість продажу цільової валюти більша за вартість базової валюти помножена на ліміт на вхід в позицію та на ліміт на позицію - переходимо в режим спекуляції
				if targetBalance*boundBid >= baseBalance*pp.pair.GetLimitOutputOfPosition() {
					pp.pair.SetStage(pair_types.OutputOfPositionStage)
					workingOutEvent <- true
					return
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.workingOutEvent
}

func NewPairObserver(
	client *binance.Client,
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
		bookTickers:      nil,
		bookTickerStream: nil,
		bookTickerEvent:  nil,
		depths:           nil,
		depthsStream:     nil,
		depthEvent:       nil,
		klines:           nil,
		klineStream:      nil,
		klineEvent:       nil,
		stop:             stop,
		degree:           degree,
		limit:            limit,
		deltaUp:          deltaUp,
		deltaDown:        deltaDown,
		priceChanges:     nil,
		priceUp:          nil,
		priceDown:        nil,
	}
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
