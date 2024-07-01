package futures_signals

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	futures_handlers "github.com/fr0ster/go-trading-utils/binance/futures/handlers"
	futures_kline "github.com/fr0ster/go-trading-utils/binance/futures/markets/kline"

	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

type (
	PairKlinesObserver struct {
		client       *futures.Client
		degree       int
		limit        int
		interval     string
		account      *futures_account.Account
		exchangeInfo *exchange_info.ExchangeInfo
		data         *kline_types.Klines
		klineEvent   chan *futures.WsKlineEvent
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
		symbol       *futures.Symbol
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
			logrus.Debugf("Futures, Create kline data for %v", pp.symbol.Symbol)
			pp.data = kline_types.New(pp.degree, pp.interval, pp.symbol.Symbol)
		}

		ticker := time.NewTicker(pp.timeOut)
		lastResponse := time.Now()
		// Запускаємо потік для отримання оновлення klines
		if pp.klineEvent == nil {
			pp.klineEvent = make(chan *futures.WsKlineEvent, 1)
			logrus.Debugf("Futures, Start stream for %v Klines", pp.symbol.Symbol)
			wsHandler := func(event *futures.WsKlineEvent) {
				lastResponse = time.Now()
				pp.klineEvent <- event
			}
			resetEvent := make(chan bool, 1)
			wsErrorHandler := func(err error) {
				resetEvent <- true
			}
			var stopC chan struct{}
			_, stopC, _ = futures.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
			go func() {
				for {
					select {
					case <-resetEvent:
						stopC <- struct{}{}
						_, stopC, _ = futures.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
					case <-ticker.C:
						if time.Since(lastResponse) > pp.timeOut {
							stopC <- struct{}{}
							_, stopC, _ = futures.WsKlineServe(pp.symbol.Symbol, pp.interval, wsHandler, wsErrorHandler)
						}
					}
				}
			}()
		}
		futures_kline.Init(pp.data, pp.client)
	}
	return pp.klineEvent
}

func eventProcess(pp *PairKlinesObserver, current_price, last_close float64, filled bool) (float64, error) {
	if last_close == 0 {
		logrus.Debugf("Futures, Initialization for %s, last price - %f, filled is %v", pp.symbol.Symbol, current_price, filled)
		last_close = current_price
	} else {
		delta := (current_price - last_close) * 100 / last_close
		if filled {
			logrus.Debugf("Futures for %s, kline is filled, Current price - %f, last price - %f, delta - %f%%",
				pp.symbol.Symbol, current_price, last_close, delta)
		} else {
			logrus.Debugf("Futures for %s, kline is not filled, Current price - %f, last price - %f, delta - %f%%",
				pp.symbol.Symbol, current_price, last_close, delta)
		}
		if delta > pp.deltaUp*100 || delta < -pp.deltaDown*100 {
			if filled {
				logrus.Debugf("Futures, kline is filled, Price for %s is changed on %f%%", pp.symbol.Symbol, delta)
			} else {
				logrus.Debugf("Futures, kline is not filled, Price for %s is changed on %f%%", pp.symbol.Symbol, delta)
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
		logrus.Debugf("Futures, Create KLine Update Signaler for %v", pp.symbol.Symbol)
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
						if val := pp.data.GetKlines().Max(); val != nil {
							current_price := utils.ConvStrToFloat64(val.(*kline_types.Kline).Close)
							last_close, err = eventProcess(pp, current_price, last_close, filled)
							if err != nil {
								logrus.Error(err)
								close(pp.stop)
								return
							}
						} else {
							logrus.Errorf("can't get Close from klines")
							close(pp.stop)
							return
						}
					} else if !filled && !pp.isFilledOnly { // Обробляемо незаповнену свічку
						// Остання ціна
						val := pp.GetKlines().GetLastKline()
						current_price := utils.ConvStrToFloat64(val.Close)
						last_close, err = eventProcess(pp, current_price, last_close, !filled)
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
		pp.filledEvent = futures_handlers.GetKlinesUpdateGuard(pp.data, pp.klineEvent, pp.isFilledOnly)
	}
	return pp.filledEvent
}

func (pp *PairKlinesObserver) GetMinQuantity(price float64) float64 {
	return utils.ConvStrToFloat64(pp.symbol.MinNotionalFilter().Notional) / price
}

func (pp *PairKlinesObserver) SetSleepingTime(sleepingTime time.Duration) {
	pp.sleepingTime = sleepingTime
}

func (pp *PairKlinesObserver) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func NewPairKlinesObserver(
	client *futures.Client,
	symbol string,
	baseSymbol string,
	targetSymbol string,
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
		sleepingTime: 1 * time.Second,
		timeOut:      1 * time.Hour,
	}
	pp.account, err = futures_account.New(pp.client, pp.degree, []string{baseSymbol}, []string{targetSymbol})
	if err != nil {
		return
	}
	pp.exchangeInfo = exchange_info.New()
	err = futures_exchange_info.Init(pp.exchangeInfo, degree, client)
	if err != nil {
		return
	}
	if symbol := pp.exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: symbol}); symbol != nil {
		pp.symbol, err = symbol.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
		if err != nil {
			logrus.Errorf(errorMsg, err)
			return
		}
	}

	return
}
