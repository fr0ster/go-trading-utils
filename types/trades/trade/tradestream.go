package trade

import (
	"time"
)

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
					// Встановлюємо новий час відповіді
					lastResponse = time.Now()
				}
			}
		}
	}()
	return
}
