package processor

// import (
// 	"time"

// 	"github.com/adshao/go-binance/v2"
// 	"github.com/sirupsen/logrus"
// )

// func (pp *PairProcessor) startBookTickerStream(handler binance.WsBookTickerHandler, errHandler binance.ErrHandler) (
// 	doneC,
// 	stopC chan struct{},
// 	err error) {
// 	// Запускаємо стрім подій користувача
// 	doneC, stopC, err = binance.WsBookTickerServe(pp.symbol.Symbol, handler, errHandler)
// 	return
// }

// func (pp *PairProcessor) BookTickerEventStart(
// 	stop chan struct{},
// 	levels int,
// 	rate time.Duration,
// 	callBack binance.WsBookTickerHandler) (
// 	resetEvent chan error,
// 	err error) {
// 	// Ініціалізуємо стріми для відмірювання часу
// 	ticker := time.NewTicker(pp.timeOut)
// 	// Ініціалізуємо маркер для останньої відповіді
// 	lastResponse := time.Now()
// 	// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
// 	resetEvent = make(chan error, 1)
// 	// Ініціалізуємо обробник помилок
// 	wsErrorHandler := func(err error) {
// 		logrus.Errorf("Future wsErrorHandler error: %v", err)
// 		resetEvent <- err
// 	}
// 	// Запускаємо стрім подій користувача
// 	_, stopC, err := pp.startBookTickerStream(callBack, wsErrorHandler)
// 	// Запускаємо стрім для перевірки часу відповіді та оновлення стріму подій користувача при необхідності
// 	go func() {
// 		for {
// 			select {
// 			case <-stop:
// 				// Зупиняємо стрім подій користувача
// 				stopC <- struct{}{}
// 				return
// 			case <-resetEvent:
// 				// Запускаємо новий стрім подій користувача
// 				_, stopC, err = pp.startBookTickerStream(callBack, wsErrorHandler)
// 				if err != nil {
// 					close(pp.stop)
// 					return
// 				}
// 			case <-ticker.C:
// 				// Перевіряємо чи не вийшли за ліміт часу відповіді
// 				if time.Since(lastResponse) > pp.timeOut {
// 					// Зупиняємо стрім подій користувача
// 					stopC <- struct{}{}
// 					// Запускаємо новий стрім подій користувача
// 					_, stopC, err = pp.startBookTickerStream(callBack, wsErrorHandler)
// 					if err != nil {
// 						close(pp.stop)
// 						return
// 					}
// 					// Встановлюємо новий час відповіді
// 					lastResponse = time.Now()
// 				}
// 			}
// 		}
// 	}()
// 	return
// }
