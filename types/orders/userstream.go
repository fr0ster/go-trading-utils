package orders

import (
	"errors"
	"time"
)

func (o *Orders) MarkStreamAsStarted() {
	o.isStartedStream = true
}

func (o *Orders) MarkStreamAsStopped() {
	o.isStartedStream = false
}

func (o *Orders) IsStreamStarted() bool {
	return o.isStartedStream
}

func (o *Orders) StreamStart() (err error) {
	// Ініціалізуємо стріми для відмірювання часу
	ticker := time.NewTicker(o.timeOut)
	// Ініціалізуємо маркер для останньої відповіді
	lastResponse := time.Now()
	// Запускаємо стрім подій користувача
	_, stopC, err := o.startUserDataStream()
	if err != nil {
		close(o.stop)
		return
	}
	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
	go func() {
		for {
			select {
			case <-o.stop:
				// Зупиняємо стрім подій користувача
				stopC <- struct{}{}
				return
			case <-o.resetEvent:
				// Запускаємо новий стрім подій користувача
				_, stopC, err = o.startUserDataStream()
				if err != nil {
					o.StreamStop()
					return
				}
				o.MarkStreamAsStarted()
			case <-ticker.C:
				// Перевіряємо чи не вийшли за ліміт часу відповіді
				if time.Since(lastResponse) > o.timeOut {
					// Зупиняємо стрім подій користувача
					stopC <- struct{}{}
					// Запускаємо новий стрім подій користувача
					_, stopC, err = o.startUserDataStream()
					if err != nil {
						o.StreamStop()
						return
					}
					o.MarkStreamAsStarted()
					// Встановлюємо новий час відповіді
					lastResponse = time.Now()
				}
			}
		}
	}()
	return
}

func (o *Orders) StreamStop() (err error) {
	if o.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(o.stop)
	o.MarkStreamAsStopped()
	return
}
