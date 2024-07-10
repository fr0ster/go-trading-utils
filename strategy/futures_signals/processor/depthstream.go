package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

const (
	DepthStreamLevel5  DepthStreamLevel = 5
	DepthStreamLevel10 DepthStreamLevel = 10
	DepthStreamLevel20 DepthStreamLevel = 20
	DepthAPILimit5     DepthAPILimit    = 5
	DepthAPILimit10    DepthAPILimit    = 10
	DepthAPILimit20    DepthAPILimit    = 20
	DepthAPILimit50    DepthAPILimit    = 50
	DepthAPILimit100   DepthAPILimit    = 100
	DepthAPILimit500   DepthAPILimit    = 500
	DepthAPILimit1000  DepthAPILimit    = 1000
	DepthStreamRate100 DepthStreamRate  = DepthStreamRate(100 * time.Millisecond)
	DepthStreamRate250 DepthStreamRate  = DepthStreamRate(250 * time.Millisecond)
	DepthStreamRate500 DepthStreamRate  = DepthStreamRate(500 * time.Millisecond)
)

type (
	DepthStreamLevel int
	DepthAPILimit    int
	DepthStreamRate  time.Duration
)

func (pp *PairProcessor) startDepthStream(
	levels DepthStreamLevel,
	rate DepthStreamRate,
	handler futures.WsDepthHandler,
	errHandler futures.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Запускаємо стрім подій користувача
	doneC, stopC, err = futures.WsPartialDepthServeWithRate(pp.symbol.Symbol, int(levels), time.Duration(rate), handler, errHandler)
	return
}

func (pp *PairProcessor) DepthEventStart(
	stop chan struct{},
	levels DepthStreamLevel,
	rate DepthStreamRate,
	callBack futures.WsDepthHandler) (
	resetEvent chan error,
	err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(pp.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
	resetEvent = make(chan error, 1)
	// Ініціалізуємо обробник помилок
	wsErrorHandler := func(err error) {
		logrus.Errorf("Future wsErrorHandler error: %v", err)
		resetEvent <- err
	}
	// Запускаємо стрім подій користувача
	_, stopC, err := pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
				if err != nil {
					close(pp.stop)
					return
				}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > pp.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
					if err != nil {
						close(pp.stop)
						return
					}
					// Встановлюємо новий час відповіді
					lastResponse = time.Now()
				}
			}
		}
	}()
	return
}

func (pp *PairProcessor) GetDepthEventCallBack(
	depthN int,
	depth *depth_types.Depth,
	summa ...*float64) futures.WsDepthHandler {
	futures_depth.Init(depth, pp.client, depthN)
	return func(event *futures.WsDepthEvent) {
		depth.Lock()         // Locking the depths
		defer depth.Unlock() // Unlocking the depths
		if event.LastUpdateID < depth.LastUpdateID {
			return
		}
		if event.LastUpdateID >= int64(depth.LastUpdateID) {
			if !(event.FirstUpdateID <= int64(depth.LastUpdateID) && event.LastUpdateID >= int64(depth.LastUpdateID)) {
				if event.PrevLastUpdateID != int64(depth.LastUpdateID) {
					futures_depth.Init(depth, pp.client, depthN)
				}
				if event.LastUpdateID < depth.LastUpdateID {
					return
				}
			}
			for _, bid := range event.Bids {
				price, quantity, err := bid.Parse()
				if err != nil {
					return
				}
				depth.UpdateBid(price, quantity)
			}
			for _, ask := range event.Asks {
				price, quantity, err := ask.Parse()
				if err != nil {
					return
				}
				depth.UpdateAsk(price, quantity)
			}
			depth.LastUpdateID = event.LastUpdateID
		}
	}
}
