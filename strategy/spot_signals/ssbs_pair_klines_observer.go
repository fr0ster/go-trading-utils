package spot_signals

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	spot_handlers "github.com/fr0ster/go-trading-utils/binance/spot/handlers"
	spot_kline "github.com/fr0ster/go-trading-utils/binance/spot/markets/kline"

	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairKlinesObserver struct {
		client       *binance.Client
		degree       int
		limit        int
		interval     string
		account      *spot_account.Account
		exchangeInfo *exchange_info.ExchangeInfo
		data         *kline_types.Klines
		klineEvent   chan *binance.WsKlineEvent
		filledEvent  chan bool
		priceChanges chan *pair_price_types.PairDelta
		priceUp      chan bool
		priceDown    chan bool
		stop         chan struct{}
		deltaUp      float64
		deltaDown    float64
		isFilledOnly bool
		sleepingTime time.Duration
		timeOut      time.Duration
		symbol       *binance.Symbol
	}
)

func (pp *PairKlinesObserver) GetKlines() *kline_types.Klines {
	return pp.data
}

func (pp *PairKlinesObserver) GetStream() chan *binance.WsKlineEvent {
	return pp.klineEvent
}

func (pp *PairKlinesObserver) StartStream() chan *binance.WsKlineEvent {
	if pp.klineEvent == nil {
		if pp.data == nil {
			logrus.Debugf("Spot, Create kline data for %v", pp.symbol.Symbol)
			pp.data = kline_types.New(degree, pp.interval, pp.symbol.Symbol)
		}

		ticker := time.NewTicker(pp.timeOut)
		lastResponse := time.Now()
		// Запускаємо потік для отримання оновлення klines
		pp.klineEvent = make(chan *binance.WsKlineEvent, 1)
		logrus.Debugf("Spot, Start stream for %v Klines", pp.symbol.Symbol)
		wsHandler := func(event *binance.WsKlineEvent) {
			lastResponse = time.Now()
			pp.klineEvent <- event
		}
		resetEvent := make(chan bool, 1)
		wsErrorHandler := func(err error) {
			resetEvent <- true
		}
		var stopC chan struct{}
		_, stopC, _ = binance.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
		go func() {
			for {
				select {
				case <-resetEvent:
					stopC <- struct{}{}
					_, stopC, _ = binance.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
				case <-ticker.C:
					if time.Since(lastResponse) > pp.timeOut {
						stopC <- struct{}{}
						_, stopC, _ = binance.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
					}
				}
			}
		}()
		spot_kline.Init(pp.data, pp.client)
	}
	return pp.klineEvent
}

func eventProcess(pp *PairKlinesObserver, current_price, last_close float64, filled bool) (float64, error) {
	if last_close == 0 {
		logrus.Debugf("Spot, Initialization for %s, last price - %f, filled is %v", pp.symbol.Symbol, current_price, filled)
		last_close = current_price
	} else {
		delta := (current_price - last_close) * 100 / last_close
		if filled {
			logrus.Debugf("Spot for %s, kline is filled, Current price - %f, last price - %f, delta - %f%%",
				pp.symbol.Symbol, current_price, last_close, delta)
		} else {
			logrus.Debugf("Spot for %s, kline is not filled, Current price - %f, last price - %f, delta - %f%%",
				pp.symbol.Symbol, current_price, last_close, delta)
		}
		if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
			if filled {
				logrus.Debugf("Spot, kline is filled, Price for %s is changed on %f%%", pp.symbol.Symbol, delta)
			} else {
				logrus.Debugf("Spot, kline is not filled, Price for %s is changed on %f%%", pp.symbol.Symbol, delta)
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
					return
				case filled := <-pp.filledEvent: // Чекаємо на заповнення свічки
					if filled {
						// Остання ціна
						val := pp.GetKlines().GetKlines().Max()
						if val == nil {
							logrus.Errorf("can't get Close from klines")
							close(pp.stop)
							return
						}
						current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
						last_close, err = eventProcess(pp, current_price, last_close, true)
						if err != nil {
							logrus.Error(err)
							close(pp.stop)
							return
						}
					} else if !filled && !pp.isFilledOnly { // Обробляемо незаповнену свічку
						// Остання ціна
						val := pp.GetKlines().GetLastKline()
						current_price := utils.ConvStrToFloat64(val.Close)
						last_close, err = eventProcess(pp, current_price, last_close, filled)
						if err != nil {
							logrus.Error(err)
							close(pp.stop)
							return
						}
					}
				}
				time.Sleep(pp.sleepingTime)
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
		pp.filledEvent = spot_handlers.GetKlinesUpdateGuard(pp.data, pp.klineEvent, pp.isFilledOnly)
	}
	return pp.filledEvent
}

func (pp *PairKlinesObserver) GetMinQuantity(price float64) float64 {
	return utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MinNotional) / price
}

func (pp *PairKlinesObserver) GetMaxQuantity(price float64) float64 {
	return utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MaxNotional) / price
}

func (pp *PairKlinesObserver) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairKlinesObserver) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func NewPairKlinesObserver(
	client *binance.Client,
	symbol string,
	degree int,
	limit int,
	interval string,
	deltaUp float64,
	deltaDown float64,
	stop chan struct{},
	isFilledOnly bool) (pp *PairKlinesObserver, err error) {
	pp = &PairKlinesObserver{
		client:       client,
		account:      nil,
		data:         nil,
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
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	pp.exchangeInfo = exchange_info.New()
	err = spot_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}
	if symbol := pp.exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: symbol}); symbol != nil {
		pp.symbol, err = symbol.(*symbol_info.SpotSymbol).GetSpotSymbol()
		if err != nil {
			logrus.Errorf(errorMsg, err)
			return
		}
	}
	pp.account, err = spot_account.New(pp.client, []string{pp.symbol.QuoteAsset, pp.symbol.BaseAsset})
	if err != nil {
		return
	}

	return
}
