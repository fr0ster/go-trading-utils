package spot_signals

import (
	_ "net/http/pprof"
	"reflect"
	"time"

	"os"

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
		client         *binance.Client
		pair           pairs_interfaces.Pairs
		degree         int
		limit          int
		interval       string
		account        *spot_account.Account
		data           *kline_types.Klines
		stream         *spot_streams.KlineStream
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

func (pp *PairKlinesObserver) GetStream() *spot_streams.KlineStream {
	return pp.stream
}

func (pp *PairKlinesObserver) StartStream() *spot_streams.KlineStream {
	if pp.stream == nil {
		if pp.data == nil {
			logrus.Debugf("Spot, Create kline data for %v", pp.pair.GetPair())
			pp.data = kline_types.New(degree, pp.interval)
		}

		// Запускаємо потік для отримання оновлення depths
		pp.stream = spot_streams.NewKlineStream(pp.pair.GetPair(), pp.interval, 1)
		logrus.Debugf("Spot, Create 2 kline goroutines for %v", pp.pair.GetPair())
		pp.stream.Start()
		spot_kline.Init(pp.data, pp.client, pp.pair.GetPair())
	}
	return pp.stream
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
					val := pp.GetKlines().GetKlines().Max()
					if val == nil {
						logrus.Error("Can't get Close from klines")
						pp.stop <- os.Interrupt
						return
					}
					current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
					delta := func() float64 { return (current_price - last_close) * 100 / last_close }
					logrus.Debugf("Spot, Current price for %s - %f, delta - %f", pp.pair.GetPair(), current_price, delta())
					logrus.Debugf("This method is implemented for %s", reflect.TypeOf(pp).Name())
					if last_close == 0 {
						last_close = current_price
					} else if current_price > last_close*(1+pp.deltaUp) {
						logrus.Debugf("Spot, Price for %s is changed on %f%%", pp.pair.GetPair(), delta())
						logrus.Debugf("This method is implemented for %s", reflect.TypeOf(pp).Name())
						pp.priceChanges <- &pair_price_types.PairDelta{
							Price: current_price, Percent: (current_price - last_close) * 100 / last_close,
						}
						last_close = current_price
						pp.priceUp <- true
					} else if current_price < last_close*(1-pp.deltaDown) {
						logrus.Debugf("Spot, Price for %s is changed on %f%%", pp.pair.GetPair(), delta())
						logrus.Debugf("This method is implemented for %s", reflect.TypeOf(pp).Name())
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
	if pp.filledEvent == nil && pp.nonFilledEvent == nil {
		if pp.stream == nil {
			pp.StartStream()
		}
		logrus.Debugf("Spot, Create Update Guard for %v", pp.pair.GetPair())
		pp.filledEvent, pp.nonFilledEvent = spot_handlers.GetKlinesUpdateGuard(pp.data, pp.stream.GetDataChannel(), pp.isFilledOnly)
	}
	return pp.filledEvent, pp.nonFilledEvent
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
