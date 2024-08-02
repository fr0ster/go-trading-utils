package depth

import (
	"errors"
	"time"
)

func (d *Depths) MarkStreamAsStarted() {
	d.isStartedStream = true
}

func (d *Depths) MarkStreamAsStopped() {
	d.isStartedStream = false
}

func (d *Depths) IsStreamStarted() bool {
	return d.isStartedStream
}

func (d *Depths) StreamStart() (err error) {
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
					d.StreamStop()
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
						d.StreamStop()
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

func (d *Depths) StreamStop() (err error) {
	if d.stop == nil {
		err = errors.New("stop channel is not initialized")
		return
	}
	close(d.stop)
	d.MarkStreamAsStopped()
	return
}
