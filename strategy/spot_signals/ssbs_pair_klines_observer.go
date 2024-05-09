package spot_signals

import (
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_kline "github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"
	"github.com/fr0ster/go-trading-utils/utils"

	spot_streams "github.com/fr0ster/go-trading-utils/binance/spot/streams"

	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairKlinesObserver struct {
		client       *binance.Client
		pair         pairs_interfaces.Pairs
		degree       int
		limit        int
		interval     string
		account      *spot_account.Account
		data         *kline_types.Klines
		stream       *spot_streams.KlineStream
		klineEvent   chan *binance.WsKlineEvent
		filledEvent  chan bool
		priceChanges chan *pair_price_types.PairDelta
		priceUp      chan bool
		priceDown    chan bool
		stop         chan os.Signal
		deltaUp      float64
		deltaDown    float64
		isFilledOnly bool
	}
)

func (pp *PairKlinesObserver) GetKlines() *kline_types.Klines {
	return pp.data
}

func (pp *PairKlinesObserver) GetStream() *spot_streams.KlineStream {
	return pp.stream
}

func (pp *PairKlinesObserver) StartStream() *spot_streams.KlineStream {
	if pp.stream == nil {
		if pp.data == nil {
			logrus.Debugf("Spot, Create kline data for %v", pp.pair.GetPair())
			pp.data = kline_types.New(degree, pp.interval, pp.pair.GetPair())
		}

		// Запускаємо потік для отримання оновлення depths
		pp.stream = spot_streams.NewKlineStream(pp.pair.GetPair(), pp.interval, 1)
		pp.stream.Start(func(event *binance.WsKlineEvent) { pp.klineEvent <- event })
		spot_kline.Init(pp.data, pp.client)
	}
	return pp.stream
}

func eventProcess(pp *PairKlinesObserver, current_price, last_close float64, filled bool) (float64, error) {
	if last_close == 0 {
		logrus.Debugf("Spot, Initialization for %s, last price - %f, %v", pp.pair.GetPair(), current_price, filled)
		last_close = current_price
	} else {
		delta := (current_price - last_close) * 100 / last_close
		if filled {
			logrus.Debugf("Spot for %s, kline is filled, Current price - %f, last price - %f, delta - %f",
				pp.pair.GetPair(), current_price, last_close, delta)
		} else {
			logrus.Debugf("Spot for %s, kline is not filled, Current price - %f, last price - %f, delta - %f",
				pp.pair.GetPair(), current_price, last_close, delta)
		}
		if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
			if filled {
				logrus.Debugf("Spot, kline is filled, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
			} else {
				logrus.Debugf("Spot, kline is not filled, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
			}
			pp.priceChanges <- &pair_price_types.PairDelta{Price: current_price, Percent: delta}
			if delta > 0 {
				pp.priceUp <- true
			} else {
				pp.priceDown <- true
			}
			last_close = current_price
		}
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
				case filled := <-pp.filledEvent: // Чекаємо на заповнення свічки
					if filled {
						// Остання ціна
						val := pp.GetKlines().GetKlines().Max()
						if val == nil {
							logrus.Errorf("can't get Close from klines")
							pp.stop <- os.Interrupt
							return
						}
						current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
						last_close, err = eventProcess(pp, current_price, last_close, true)
						if err != nil {
							logrus.Error(err)
							pp.stop <- os.Interrupt
							return
						}
					} else if !filled && !pp.isFilledOnly { // Обробляемо незаповнену свічку
						// Остання ціна
						val := pp.GetKlines().GetLastKline()
						current_price := utils.ConvStrToFloat64(val.Close)
						last_close, err = eventProcess(pp, current_price, last_close, filled)
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

func (pp *PairKlinesObserver) StartUpdateGuard() chan bool {
	if pp.filledEvent == nil {
		if pp.stream == nil {
			pp.StartStream()
		}
		pp.filledEvent = spot_handlers.GetKlinesUpdateGuard(pp.data, pp.klineEvent, pp.isFilledOnly)
	}
	return pp.filledEvent
}

func NewPairKlinesObserver(
	client *binance.Client,
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
	pp.account, err = spot_account.New(pp.client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	return
}
