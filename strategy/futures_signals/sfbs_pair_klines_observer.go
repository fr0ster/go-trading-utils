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

	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
)

type (
	PairKlinesObserver struct {
		client       *futures.Client
		pair         pairs_interfaces.Pairs
		degree       int
		limit        int
		interval     string
		account      *futures_account.Account
		data         *kline_types.Klines
		klineEvent   chan *futures.WsKlineEvent
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

func (pp *PairKlinesObserver) GetStream() chan *futures.WsKlineEvent {
	return pp.klineEvent
}

func (pp *PairKlinesObserver) StartStream() chan *futures.WsKlineEvent {
	if pp.klineEvent == nil {
		if pp.data == nil {
			logrus.Debugf("Futures, Create kline data for %v", pp.pair.GetPair())
			pp.data = kline_types.New(pp.degree, pp.interval, pp.pair.GetPair())
		}
		// Запускаємо потік для отримання оновлення klines
		if pp.klineEvent == nil {
			pp.klineEvent = make(chan *futures.WsKlineEvent, 1)
			logrus.Debugf("Futures, Start stream for %v Klines", pp.pair.GetPair())
			wsHandler := func(event *futures.WsKlineEvent) {
				pp.klineEvent <- event
			}
			futures.WsKlineServe(pp.pair.GetPair(), pp.interval, wsHandler, utils.HandleErr)
		}
		futures_kline.Init(pp.data, pp.client)
	}
	return pp.klineEvent
}

func eventProcess(pp *PairKlinesObserver, current_price, last_close float64, filled bool) (float64, error) {
	if last_close == 0 {
		logrus.Debugf("Futures, Initialization for %s, last price - %f, filled is %v", pp.pair.GetPair(), current_price, filled)
		last_close = current_price
	} else {
		delta := (current_price - last_close) * 100 / last_close
		if filled {
			logrus.Debugf("Futures for %s, kline is filled, Current price - %f, last price - %f, delta - %f%%",
				pp.pair.GetPair(), current_price, last_close, delta)
		} else {
			logrus.Debugf("Futures for %s, kline is not filled, Current price - %f, last price - %f, delta - %f%%",
				pp.pair.GetPair(), current_price, last_close, delta)
		}
		if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
			if filled {
				logrus.Debugf("Futures, kline is filled, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
			} else {
				logrus.Debugf("Futures, kline is not filled, Price for %s is changed on %f%%", pp.pair.GetPair(), delta)
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
		logrus.Debugf("Futures, Create KLine Update Signaler for %v", pp.pair.GetPair())
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
						if val := pp.data.GetKlines().Max(); val != nil {
							current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
							last_close, err = eventProcess(pp, current_price, last_close, filled)
							if err != nil {
								logrus.Error(err)
								pp.stop <- os.Interrupt
								return
							}
						} else {
							logrus.Errorf("can't get Close from klines")
							pp.stop <- os.Interrupt
							return
						}
					} else if !filled && !pp.isFilledOnly { // Обробляемо незаповнену свічку
						// Остання ціна
						val := pp.GetKlines().GetLastKline()
						current_price := utils.ConvStrToFloat64(val.Close)
						last_close, err = eventProcess(pp, current_price, last_close, !filled)
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
		if pp.klineEvent == nil {
			pp.StartStream()
		}
		pp.filledEvent = futures_handlers.GetKlinesUpdateGuard(pp.data, pp.klineEvent, pp.isFilledOnly)
	}
	return pp.filledEvent
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
		klineEvent:   nil,
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
