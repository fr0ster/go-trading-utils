package processor

// import (
// 	"time"

// 	"github.com/adshao/go-binance/v2/futures"
// 	"github.com/sirupsen/logrus"

// 	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
// 	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
// 	depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
// 	types "github.com/fr0ster/go-trading-utils/types/depths/items"
// )

// func (pp *PairProcessor) startDepthStream(
// 	levels depths_types.DepthStreamLevel,
// 	rate depths_types.DepthStreamRate,
// 	handler futures.WsDepthHandler,
// 	errHandler futures.ErrHandler) (
// 	doneC,
// 	stopC chan struct{},
// 	err error) {
// 	// Запускаємо стрім подій користувача
// 	doneC, stopC, err = futures.WsPartialDepthServeWithRate(pp.symbol.Symbol, int(levels), time.Duration(rate), handler, errHandler)
// 	return
// }

// func (pp *PairProcessor) DepthEventStart(
// 	stop chan struct{},
// 	levels depths_types.DepthStreamLevel,
// 	rate depths_types.DepthStreamRate,
// 	callBack futures.WsDepthHandler) (
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
// 	_, stopC, err := pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
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
// 				_, stopC, err = pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
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
// 					_, stopC, err = pp.startDepthStream(levels, rate, callBack, wsErrorHandler)
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

// func (pp *PairProcessor) GetDepth() *depth_types.Depths {
// 	return pp.depth
// }

// func (pp *PairProcessor) GetDepthEventCallBack() futures.WsDepthHandler {
// 	futures_depth.Init(pp.depth, pp.client)
// 	return func(event *futures.WsDepthEvent) {
// 		pp.depth.Lock()         // Locking the depths
// 		defer pp.depth.Unlock() // Unlocking the depths
// 		if event.LastUpdateID < pp.depth.LastUpdateID {
// 			return
// 		}
// 		if event.PrevLastUpdateID != int64(pp.depth.LastUpdateID) {
// 			futures_depth.Init(pp.depth, pp.client)
// 		} else if event.PrevLastUpdateID == int64(pp.depth.LastUpdateID) {
// 			for _, bid := range event.Bids {
// 				price, quantity, err := bid.Parse()
// 				if err != nil {
// 					return
// 				}
// 				pp.depth.GetBids().Update(types.NewBid(types.PriceType(price), types.QuantityType(quantity)))
// 			}
// 			for _, ask := range event.Asks {
// 				price, quantity, err := ask.Parse()
// 				if err != nil {
// 					return
// 				}
// 				pp.depth.GetAsks().Update(types.NewAsk(types.PriceType(price), types.QuantityType(quantity)))
// 			}
// 			pp.depth.LastUpdateID = event.LastUpdateID
// 		}
// 	}
// }
