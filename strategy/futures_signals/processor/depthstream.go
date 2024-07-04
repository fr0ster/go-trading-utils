package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"
)

const (
	DepthStreamLevel5   DepthStreamLevel = 5
	DepthStreamLevel10  DepthStreamLevel = 10
	DepthStreamLevel20  DepthStreamLevel = 20
	DepthStreamLevel100 DepthStreamLevel = 100
	DepthStreamLevel250 DepthStreamLevel = 250
	DepthStreamLevel500 DepthStreamLevel = 500
)

type (
	DepthStreamLevel int
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
