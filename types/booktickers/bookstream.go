package booktickers

import (
	"errors"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

func (bt *BookTickers) MarkStreamAsStarted() {
	bt.isStartedStream = true
}

func (bt *BookTickers) MarkStreamAsStopped() {
	bt.isStartedStream = false
}

func (bt *BookTickers) IsStreamStarted() bool {
	return bt.isStartedStream
}

func (bt *BookTickers) StreamStart(
	levels int,
	rate time.Duration,
	callBack futures.WsBookTickerHandler) (err error) {
	if bt.init == nil || bt.startBookTickerStream == nil {
		err = errors.New("initial functions for Streams and Data are not initialized")
		return
	}
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(bt.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := bt.startBookTickerStream()
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-bt.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-bt.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = bt.startBookTickerStream()
				if err != nil {
					bt.StreamStop()
					return
				}
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > bt.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = bt.startBookTickerStream()
					if err != nil {
						bt.StreamStop()
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

func (bt *BookTickers) StreamStop() (err error) {
	if bt.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(bt.stop)
	bt.MarkStreamAsStopped()
	return
}
