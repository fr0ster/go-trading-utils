package depth

import (
	"errors"
	"time"
)

func (d *Depths) DepthEventStart() (err error) {
	if d.Init == nil || d.startDepthStream == nil {
		err = errors.New("initial functions for Streams and Data are not initialized")
		return
	}
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(d.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := d.startDepthStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-d.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-d.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = d.startDepthStream()
				if err != nil {
					close(d.stop)
					return
				}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > d.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = d.startDepthStream()
					if err != nil {
						close(d.stop)
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
