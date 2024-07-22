package kline

import (
	"errors"
	"time"
)

const (
	// Kline interval
	// KlineStreamInterval1s  KlineStreamInterval = "1s"
	KlineStreamInterval1m  KlineStreamInterval = "1m"
	KlineStreamInterval3m  KlineStreamInterval = "3m"
	KlineStreamInterval5m  KlineStreamInterval = "5m"
	KlineStreamInterval15m KlineStreamInterval = "15m"
	KlineStreamInterval30m KlineStreamInterval = "30m"
	KlineStreamInterval1h  KlineStreamInterval = "1h"
	KlineStreamInterval2h  KlineStreamInterval = "2h"
	KlineStreamInterval4h  KlineStreamInterval = "4h"
	KlineStreamInterval6h  KlineStreamInterval = "6h"
	KlineStreamInterval8h  KlineStreamInterval = "8h"
	KlineStreamInterval12h KlineStreamInterval = "12h"
	KlineStreamInterval1d  KlineStreamInterval = "1d"
	KlineStreamInterval3d  KlineStreamInterval = "3d"
	KlineStreamInterval1w  KlineStreamInterval = "1w"
	KlineStreamInterval1M  KlineStreamInterval = "1M"
)

type (
	KlineStreamInterval string
)

func (pp *Klines) KlineEventStart() (err error) {
	if pp.init == nil || pp.startKlineStream == nil {
		err = errors.New("initial functions for Streams and Data are not initialized")
		return
	}
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(pp.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := pp.startKlineStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-pp.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-pp.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = pp.startKlineStream()
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
					_, stopC, err = pp.startKlineStream()
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
