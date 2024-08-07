package aggtrade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	aggtrade_types "github.com/fr0ster/go-trading-utils/types/trades/aggtrade"
	"github.com/sirupsen/logrus"
)

func InitCreator(client *futures.Client, limit int) func(at *aggtrade_types.AggTrades) types.InitFunction {
	return func(at *aggtrade_types.AggTrades) types.InitFunction {
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

func CallBackCreator(limit int) func(trade *aggtrade_types.AggTrades) futures.WsAggTradeHandler {
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
func TradeStreamCreator(
	handler func(trade *aggtrade_types.AggTrades) futures.WsAggTradeHandler,
	errHandler func(trade *aggtrade_types.AggTrades) futures.ErrHandler) func(*aggtrade_types.AggTrades) types.StreamFunction {
	return func(at *aggtrade_types.AggTrades) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsAggTradeServe(at.Symbol(), handler(at), errHandler(at))
			return
		}
	}
}

func WsErrorHandlerCreator() func(*aggtrade_types.AggTrades) futures.ErrHandler {
	return func(at *aggtrade_types.AggTrades) futures.ErrHandler {
		return func(err error) {
			logrus.Errorf("Future AggTrade error: %v", err)
			at.ResetEvent(err)
		}
	}
}
