package aggtrade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	aggtrade_types "github.com/fr0ster/go-trading-utils/types/trades/aggtrade"
)

func GetAggTradeInitCreator(client *futures.Client, limit int) func(at *aggtrade_types.AggTrades) func() (err error) {
	return func(at *aggtrade_types.AggTrades) func() (err error) {
		return func() (err error) {
			res, err :=
				client.NewAggTradesService().
					Symbol(string(at.Symbol())).
					Limit(limit).
					Do(context.Background())
			if err != nil {
				return err
			}
			for _, trade := range res {
				at.Update(&aggtrade_types.AggTrade{
					AggTradeID:   trade.AggTradeID,
					Price:        trade.Price,
					Quantity:     trade.Quantity,
					FirstTradeID: trade.FirstTradeID,
					LastTradeID:  trade.LastTradeID,
					Timestamp:    trade.Timestamp,
					IsBuyerMaker: trade.IsBuyerMaker,
					// IsBestPriceMatch: trade.IsBestPriceMatch,
				})
			}
			return nil
		}
	}
}

func StandardEventCallBackCreator(limit int) func(trade *aggtrade_types.AggTrades) futures.WsAggTradeHandler {
	return func(trade *aggtrade_types.AggTrades) futures.WsAggTradeHandler {
		return func(event *futures.WsAggTradeEvent) {
			trade.Lock()         // Locking the depths
			defer trade.Unlock() // Unlocking the depths
			trade.Update(&aggtrade_types.AggTrade{
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
}
func GetStartTradeStreamCreator(
	handler func(trade *aggtrade_types.AggTrades) futures.WsAggTradeHandler,
	errHandler func(trade *aggtrade_types.AggTrades) futures.ErrHandler) func(*aggtrade_types.AggTrades) func() (doneC, stopC chan struct{}, err error) {
	return func(at *aggtrade_types.AggTrades) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsAggTradeServe(at.Symbol(), handler(at), errHandler(at))
			return
		}
	}
}
