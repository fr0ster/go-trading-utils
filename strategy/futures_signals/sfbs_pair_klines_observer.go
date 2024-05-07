package futures_signals

import (
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_kline "github.com/fr0ster/go-trading-utils/binance/futures/markets/kline"
	"github.com/fr0ster/go-trading-utils/utils"

	futures_streams "github.com/fr0ster/go-trading-utils/binance/futures/streams"

	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairKlinesObserver struct {
		client         *futures.Client
		pair           pairs_interfaces.Pairs
		degree         int
		limit          int
		interval       string
		account        *futures_account.Account
		data           *kline_types.Klines
		stream         *futures_streams.KlineStream
		filledEvent    chan bool
		nonFilledEvent chan bool
		priceChanges   chan *pair_price_types.PairDelta
		priceUp        chan bool
		priceDown      chan bool
		stop           chan os.Signal
		deltaUp        float64
		deltaDown      float64
		isFilledOnly   bool
	}
)

func (pp *PairKlinesObserver) GetKlines() *kline_types.Klines {
	return pp.data
}

func (pp *PairKlinesObserver) GetStream() *futures_streams.KlineStream {
	return pp.stream
}

func (pp *PairKlinesObserver) StartStream() *futures_streams.KlineStream {
	if pp.stream == nil {
		if pp.data == nil {
			logrus.Debugf("Futures, Create kline data for %v", pp.pair.GetPair())
			pp.data = kline_types.New(pp.degree, pp.interval)
		}

		// Запускаємо потік для отримання оновлення depths
		pp.stream = futures_streams.NewKlineStream(pp.pair.GetPair(), pp.interval, 1)
		logrus.Debugf("Futures, Create 2 goroutine for %v Klines", pp.pair.GetPair())
		pp.stream.Start()
		futures_kline.Init(pp.data, pp.client, pp.pair.GetPair())
	}
	return pp.stream
}

func delta(current_price, last_close float64) float64 {
	return (current_price - last_close) * 100 / last_close
}
func eventProcess(pp *PairKlinesObserver, current_price, last_close float64) (float64, error) {
	if last_close != 0 {
		logrus.Debugf("Spot for %s, Current price - %f, last price - %f, delta - %f",
			pp.pair.GetPair(), current_price, last_close, delta(current_price, last_close))
	}
	if last_close == 0 {
		last_close = current_price
	}
	delta := delta(current_price, last_close)
	if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
		logrus.Debugf("Spot, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
		pp.priceChanges <- &pair_price_types.PairDelta{Price: current_price, Percent: delta}
		if delta > 0 {
			pp.priceUp <- true
		} else {
			pp.priceDown <- true
		}
		last_close = current_price
	}
	return last_close, nil
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
			var (
				last_close float64
				err        error
			)
			for {
				select {
				case <-pp.stop:
					pp.stop <- os.Interrupt
					return
				case <-pp.filledEvent: // Чекаємо на заповнення свічки
					// Остання ціна
					val := pp.GetKlines().GetKlines().Max()
					if val == nil {
						logrus.Errorf("can't get Close from klines")
						pp.stop <- os.Interrupt
						return
					}
					current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
					last_close, err = eventProcess(pp, current_price, last_close)
					if err != nil {
						logrus.Error(err)
						pp.stop <- os.Interrupt
						return
					}
				case <-pp.nonFilledEvent: // Обробляемо незаповнену свічку
					if !pp.isFilledOnly {
						// Остання ціна
						val := pp.GetKlines().GetLastKline()
						current_price := utils.ConvStrToFloat64(val.Close)
						last_close, err = eventProcess(pp, current_price, last_close)
						if err != nil {
							logrus.Error(err)
							pp.stop <- os.Interrupt
							return
						}
					}
				}
				time.Sleep(pp.pair.GetSleepingTime())
			}
		}()
	}
	return pp.priceChanges, pp.priceUp, pp.priceDown
}

func (pp *PairKlinesObserver) StartUpdateGuard() (chan bool, chan bool) {
	if pp.filledEvent == nil && pp.nonFilledEvent == nil {
		if pp.stream == nil {
			pp.StartStream()
		}
		logrus.Debugf("Futures, Create Update Guard for %v", pp.pair.GetPair())
		pp.filledEvent, pp.nonFilledEvent = futures_handlers.GetKlinesUpdateGuard(pp.data, pp.stream.GetDataChannel(), pp.isFilledOnly)
	}
	return pp.filledEvent, pp.nonFilledEvent
}

func NewPairKlinesObserver(
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	degree int,
	limit int,
	interval string,
	deltaUp float64,
	deltaDown float64,
	stop chan os.Signal,
	isFilledOnly bool) (pp *PairKlinesObserver, err error) {
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
		interval:     interval,
		deltaUp:      deltaUp,
		deltaDown:    deltaDown,
		priceChanges: nil,
		priceUp:      nil,
		priceDown:    nil,
		isFilledOnly: isFilledOnly,
	}
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
