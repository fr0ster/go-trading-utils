package tradeV3

import (
	"errors"
	"time"
)

func (tv3 *TradesV3) MarkStreamAsStarted() {
	tv3.isStartedStream = true
}

func (tv3 *TradesV3) MarkStreamAsStopped() {
	tv3.isStartedStream = false
}

func (tv3 *TradesV3) IsStreamStarted() bool {
	return tv3.isStartedStream
}

func (tv3 *TradesV3) StreamStart() (err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(tv3.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := tv3.startTradeStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-tv3.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-tv3.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = tv3.startTradeStream()
				if err != nil {
					close(tv3.stop)
					return
				}
				tv3.MarkStreamAsStarted()
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > tv3.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = tv3.startTradeStream()
					if err != nil {
						close(tv3.stop)
						return
					}
					tv3.MarkStreamAsStarted()
					// Встановлюємо новий час відповіді
					lastResponse = time.Now()
				}
			}
		}
	}()
	return
}

func (tv3 *TradesV3) StreamStop() (err error) {
	if tv3.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(tv3.stop)
	tv3.MarkStreamAsStopped()
	return
}
