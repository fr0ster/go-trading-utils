package processor

import (
	"context"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"
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

func (pp *PairProcessor) startKlineStream(interval KlineStreamInterval, handler futures.WsKlineHandler, errHandler futures.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Запускаємо стрім подій користувача
	doneC, stopC, err = futures.WsKlineServe(pp.symbol.Symbol, string(interval), handler, errHandler)
	return
}

func (pp *PairProcessor) KlineEventStart(
	stop chan struct{},
	interval KlineStreamInterval,
	callBack futures.WsKlineHandler) (
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
	_, stopC, err := pp.startKlineStream(interval, callBack, wsErrorHandler)
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
				_, stopC, err = pp.startKlineStream(interval, callBack, wsErrorHandler)
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
					_, stopC, err = pp.startKlineStream(interval, callBack, wsErrorHandler)
					if err != nil {
						close(pp.stop)
						return
					}
				}
			}
		}
	}()
	return
}

func (pp *PairProcessor) GetKlines(interval KlineStreamInterval, limit int) ([]*futures.Kline, error) {
	return pp.client.NewKlinesService().Symbol(pp.symbol.Symbol).Interval(string(interval)).Limit(limit).Do(context.Background())
}

// Функція для обчислення коефіцієнтів a та b за методом найменших квадратів
func (pp *PairProcessor) LeastSquares(y []float64) (a float64, b float64) {
	var sumX, sumY, sumXY, sumX2 float64
	x := []float64{}
	N := float64(len(x))

	for i := 0; i < int(N); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	a = (N*sumXY - sumX*sumY) / (N*sumX2 - sumX*sumX)
	b = (sumY - a*sumX) / N

	return a, b
}

// Функція для обчислення кута нахилу тренду з коефіцієнта нахилу a
func (pp *PairProcessor) CalculateTrendAngle(a float64) float64 {
	return math.Atan(a) * (180 / math.Pi) // Перетворення радіанів в градуси
}

func (pp *PairProcessor) InitKlinesBuffer(size int) {
	pp.klinesBuffer = make([]float64, size)
}

// Функція для додавання нового значення до кільцевого буфера
func (pp *PairProcessor) AddToBuffer(value float64) []float64 {
	pp.klinesBuffer = append(pp.klinesBuffer[1:], value) // Видаляємо перший елемент і додаємо новий в кінець
	return pp.klinesBuffer
}
