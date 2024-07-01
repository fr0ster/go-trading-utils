package processor

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"
)

func (pp *PairProcessor) startUserDataStream(handler futures.WsUserDataHandler, errHandler futures.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Отримуємо новий або той же самий ключ для прослуховування подій користувача при втраті з'єднання
	listenKey, err := pp.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	// Запускаємо стрім подій користувача
	doneC, stopC, err = futures.WsUserDataServe(listenKey, handler, errHandler)
	return
}

func (pp *PairProcessor) UserDataEventStart(
	stop chan struct{},
	callBack futures.WsUserDataHandler,
	eventType ...futures.UserDataEventType) (
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
	// Ініціалізуємо обробник подій
	eventMap := make(map[futures.UserDataEventType]bool)
	for _, event := range eventType {
		eventMap[event] = true
	}
	wsHandler := func(event *futures.WsUserDataEvent) {
		if len(eventType) == 0 || eventMap[event.Event] {
			callBack(event)
		}
	}
	// Запускаємо стрім подій користувача
	_, stopC, err := pp.startUserDataStream(wsHandler, wsErrorHandler)
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				close(pp.stop)
				return
			case <-resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = pp.startUserDataStream(wsHandler, wsErrorHandler)
				if err != nil {
					close(pp.stop)
					return
				}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > pp.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
				}
				// Запускаємо новий стрім подій користувача
				_, stopC, err = pp.startUserDataStream(wsHandler, wsErrorHandler)
				if err != nil {
					close(pp.stop)
					return
				}
			}
		}
	}()
	return
}
