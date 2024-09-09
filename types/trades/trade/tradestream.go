package trade

import (
	"errors"
	"time"
)

func (t *Trades) MarkStreamAsStarted() {
	t.isStartedStream = true
}

func (t *Trades) MarkStreamAsStopped() {
	t.isStartedStream = false
}

func (t *Trades) IsStreamStarted() bool {
	return t.isStartedStream
}

func (t *Trades) StreamStart() (err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(t.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := t.startTradeStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-t.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-t.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = t.startTradeStream()
				if err != nil {
					close(t.stop)
					return
				}
				t.MarkStreamAsStarted()
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > t.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = t.startTradeStream()
					if err != nil {
						close(t.stop)
						return
					}
					t.MarkStreamAsStarted()
					// Встановлюємо новий час відповіді
					lastResponse = time.Now()
				}
			}
		}
	}()
	return
}

func (t *Trades) StreamStop() (err error) {
	if t.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(t.stop)
	t.MarkStreamAsStopped()
	return
}
