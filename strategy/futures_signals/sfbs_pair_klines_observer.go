package futures_signals

import (
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_kline "github.com/fr0ster/go-trading-utils/binance/futures/markets/kline"
	"github.com/fr0ster/go-trading-utils/utils"

	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	pair_types "github.com/fr0ster/go-trading-utils/types/pairs"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairKlinesObserver struct {
		client             *futures.Client
		pair               pairs_interfaces.Pairs
		degree             int
		limit              int
		account            *futures_account.Account
		data               *kline_types.Klines
		stream             *futures_streams.KlineStream
		filledEvent        chan bool
		nonFilledEvent     chan bool
		collectionOutEvent chan bool
		workingOutEvent    chan bool
		priceChanges       chan *pair_price_types.PairDelta
		priceUp            chan bool
		priceDown          chan bool
		stop               chan os.Signal
		deltaUp            float64
		deltaDown          float64
	}
)

func (pp *PairKlinesObserver) Get() *kline_types.Klines {
	return pp.data
}

func (pp *PairKlinesObserver) GetStream() *futures_streams.KlineStream {
	return pp.stream
}

func (pp *PairKlinesObserver) StartStream() *futures_streams.KlineStream {
	if pp.stream == nil {
		if pp.data == nil {
			pp.data = kline_types.New(degree)
		}

		// Запускаємо потік для отримання оновлення depths
		pp.stream = futures_streams.NewKlineStream(pp.pair.GetPair(), "1m", 1)
		pp.stream.Start()
		futures_kline.Init(pp.data, pp.client, pp.pair.GetPair())
	}
	return pp.stream
}

func (pp *PairKlinesObserver) StartWorkInPositionSignal(triggerEvent chan bool) (
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

func (pp *PairKlinesObserver) StopWorkInPositionSignal(triggerEvent chan bool) (
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

func (pp *PairKlinesObserver) StartPriceChangesSignal() (
	priceChanges chan *pair_price_types.PairDelta,
	priceUp chan bool,
	priceDown chan bool) {
	if pp.priceChanges == nil && pp.priceUp == nil && pp.priceDown == nil {
		pp.priceChanges = make(chan *pair_price_types.PairDelta, 1)
		pp.priceUp = make(chan bool, 1)
		pp.priceDown = make(chan bool, 1)
		go func() {
			var last_close float64
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-pp.filledEvent: // Чекаємо на спрацювання тригера на зміну ціни
					// Остання ціна
					val := pp.Get().GetKlines().Max()
					if val == nil {
						logrus.Error("Can't get Close from klines")
						pp.stop <- os.Interrupt
						return
					}
					current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
					if last_close == 0 {
						last_close = current_price
					} else if last_close > current_price*(1+pp.deltaUp) {
						pp.priceChanges <- &pair_price_types.PairDelta{
							Price: current_price, Percent: (current_price - last_close) * 100 / last_close,
						}
						last_close = current_price
						pp.priceUp <- true
					} else if last_close < current_price*(1-pp.deltaDown) {
						pp.priceChanges <- &pair_price_types.PairDelta{
							Price: current_price, Percent: (current_price - last_close) * 100 / last_close,
						}
						last_close = current_price
						pp.priceDown <- true
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.priceChanges, pp.priceUp, pp.priceDown
}

func (pp *PairKlinesObserver) StartUpdateGuard() (chan bool, chan bool) {
	if pp.filledEvent == nil || pp.nonFilledEvent == nil {
		if pp.stream == nil {
			pp.StartStream()
		}
		pp.filledEvent, pp.nonFilledEvent = futures_handlers.GetKlinesUpdateGuard(pp.data, pp.stream.GetDataChannel(), true)
	}
	return pp.filledEvent, pp.nonFilledEvent
}

func NewPairKlinesObserver(
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal) (pp *PairKlinesObserver, err error) {
	pp = &PairKlinesObserver{
		client:       client,
		pair:         pair,
		account:      nil,
		data:         nil,
		stream:       nil,
		filledEvent:  nil,
		stop:         stop,
		degree:       degree,
		limit:        limit,
		deltaUp:      deltaUp,
		deltaDown:    deltaDown,
		priceChanges: nil,
		priceUp:      nil,
		priceDown:    nil,
	}
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
