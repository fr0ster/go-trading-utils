package processor

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	binance_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
)

const (
	DepthStreamLevel5  DepthStreamLevel = "5"
	DepthStreamLevel10 DepthStreamLevel = "10"
	DepthStreamLevel20 DepthStreamLevel = "20"
	DepthAPILimit5     DepthAPILimit    = "5"
	DepthAPILimit10    DepthAPILimit    = "10"
	DepthAPILimit20    DepthAPILimit    = "20"
	DepthAPILimit50    DepthAPILimit    = "50"
	DepthAPILimit100   DepthAPILimit    = "100"
	DepthAPILimit500   DepthAPILimit    = "500"
	DepthAPILimit1000  DepthAPILimit    = "1000"
)

type (
	DepthStreamLevel string
	DepthAPILimit    string
	DepthStreamRate  time.Duration
)

func (pp *PairProcessor) startDepthStream(
	handler binance.WsDepthHandler,
	errHandler binance.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Запускаємо стрім подій користувача
	doneC, stopC, err = binance.WsDepthServe100Ms(pp.symbol.Symbol, handler, errHandler)
	return
}

func (pp *PairProcessor) DepthEventStart(
	stop chan struct{},
	callBack binance.WsDepthHandler) (
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
	_, stopC, err := pp.startDepthStream(callBack, wsErrorHandler)
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
				_, stopC, err = pp.startDepthStream(callBack, wsErrorHandler)
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
					_, stopC, err = pp.startDepthStream(callBack, wsErrorHandler)
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

func (pp *PairProcessor) GetDepthEventCallBack(depth *depth_types.Depth) binance.WsDepthHandler {
	binance_depth.Init(depth, pp.client)
	return func(event *binance.WsDepthEvent) {
		depth.Lock()         // Locking the depths
		defer depth.Unlock() // Unlocking the depths
		if event.LastUpdateID < depth.LastUpdateID {
			return
		}
		if event.LastUpdateID != int64(depth.LastUpdateID)+1 {
			binance_depth.Init(depth, pp.client)
		} else if event.LastUpdateID == int64(depth.LastUpdateID)+1 {
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

func (pp *PairProcessor) startPartialDepthStream(
	levels DepthStreamLevel,
	handler binance.WsPartialDepthHandler,
	errHandler binance.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Запускаємо стрім подій користувача
	doneC, stopC, err = binance.WsPartialDepthServe100Ms(pp.symbol.Symbol, string(levels), handler, errHandler)
	return
}

func (pp *PairProcessor) PartialDepthEventStart(
	stop chan struct{},
	levels DepthStreamLevel,
	callBack binance.WsPartialDepthHandler) (
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
	_, stopC, err := pp.startPartialDepthStream(levels, callBack, wsErrorHandler)
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
				_, stopC, err = pp.startPartialDepthStream(levels, callBack, wsErrorHandler)
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
					_, stopC, err = pp.startPartialDepthStream(levels, callBack, wsErrorHandler)
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

func (pp *PairProcessor) GetPartialDepthEventCallBack(depth *depth_types.Depth) binance.WsPartialDepthHandler {
	binance_depth.Init(depth, pp.client)
	return func(event *binance.WsPartialDepthEvent) {
		depth.Lock()         // Locking the depths
		defer depth.Unlock() // Unlocking the depths
		if event.LastUpdateID < depth.LastUpdateID {
			return
		}
		if event.LastUpdateID != int64(depth.LastUpdateID)+1 {
			binance_depth.Init(depth, pp.client)
		} else if event.LastUpdateID == int64(depth.LastUpdateID)+1 {
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

func (pp *PairProcessor) GetDepth() *depth_types.Depth {
	return pp.depth
}

func (pp *PairProcessor) SetDepth(depth *depth_types.Depth) {
	pp.depth = depth
}
