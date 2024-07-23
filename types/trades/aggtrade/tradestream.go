package aggtrade

import (
	"time"
)

func (at *AggTrades) TradeEventStart() (err error) {
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
					close(at.stop)
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
						close(at.stop)
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
