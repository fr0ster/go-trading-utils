package aggtrade

import (
	"errors"
	"time"
)

func (at *AggTrades) MarkStreamAsStarted() {
	at.isStartedStream = true
}

func (at *AggTrades) MarkStreamAsStopped() {
	at.isStartedStream = false
}

func (at *AggTrades) IsStreamStarted() bool {
	return at.isStartedStream
}

func (at *AggTrades) StreamStart() (err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(at.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := at.startTradeStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-at.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-at.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = at.startTradeStream()
				if err != nil {
					at.StreamStop()
					return
				}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > at.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = at.startTradeStream()
					if err != nil {
						at.StreamStop()
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

func (at *AggTrades) StreamStop() (err error) {
	if at.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(at.stop)
	at.MarkStreamAsStopped()
	return
}
