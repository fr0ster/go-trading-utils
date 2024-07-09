package processor

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

// WsAggTradeServe

func (pp *PairProcessor) startTradeStream(
	handler futures.WsAggTradeHandler,
	errHandler futures.ErrHandler) (
	doneC,
	stopC chan struct{},
	err error) {
	// Запускаємо стрім подій користувача
	doneC, stopC, err = futures.WsAggTradeServe(pp.symbol.Symbol, handler, errHandler)
	return
}

func (pp *PairProcessor) TradeEventStart(
	stop chan struct{},
	callBack futures.WsAggTradeHandler) (
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
	_, stopC, err := pp.startTradeStream(callBack, wsErrorHandler)
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
				_, stopC, err = pp.startTradeStream(callBack, wsErrorHandler)
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
					_, stopC, err = pp.startTradeStream(callBack, wsErrorHandler)
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

func (pp *PairProcessor) GetAggTradesHandler(trade *trade_types.AggTrades) futures.WsAggTradeHandler {
	return func(event *futures.WsAggTradeEvent) {
		trade.Lock()         // Locking the depths
		defer trade.Unlock() // Unlocking the depths
		trade.Update(&trade_types.AggTrade{
			AggTradeID:   event.AggregateTradeID,
			Price:        event.Price,
			Quantity:     event.Quantity,
			FirstTradeID: event.FirstTradeID,
			LastTradeID:  event.LastTradeID,
			Timestamp:    event.TradeTime,
			IsBuyerMaker: event.Maker,
			// IsBestPriceMatch: event.IsBestPriceMatch,
		})
	}
}
