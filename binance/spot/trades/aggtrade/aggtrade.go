package aggtrade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	aggtrade_types "github.com/fr0ster/go-trading-utils/types/trades/aggtrade"
	"github.com/sirupsen/logrus"
)

func InitCreator(client *binance.Client, limit int) func(at *aggtrade_types.AggTrades) types.InitFunction {
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
					AggTradeID:       trade.AggTradeID,
					Price:            trade.Price,
					Quantity:         trade.Quantity,
					FirstTradeID:     trade.FirstTradeID,
					LastTradeID:      trade.LastTradeID,
					Timestamp:        trade.Timestamp,
					IsBuyerMaker:     trade.IsBuyerMaker,
					IsBestPriceMatch: trade.IsBestPriceMatch,
				})
			}
			return nil
		}
	}
}

func CallBackCreator(limit int) func(trade *aggtrade_types.AggTrades) binance.WsAggTradeHandler {
	return func(trade *aggtrade_types.AggTrades) binance.WsAggTradeHandler {
		return func(event *binance.WsAggTradeEvent) {
			trade.Lock()         // Locking the depths
			defer trade.Unlock() // Unlocking the depths
			trade.Update(&aggtrade_types.AggTrade{
				AggTradeID:   event.AggTradeID,
				Price:        event.Price,
				Quantity:     event.Quantity,
				FirstTradeID: event.FirstBreakdownTradeID,
				LastTradeID:  event.LastBreakdownTradeID,
				Timestamp:    event.TradeTime,
				IsBuyerMaker: event.IsBuyerMaker,
				// IsBestPriceMatch: event.IsBestPriceMatch,
			})
		}
	}
}
func TradeStreamCreator(
	handler func(trade *aggtrade_types.AggTrades) binance.WsAggTradeHandler,
	errHandler func(trade *aggtrade_types.AggTrades) binance.ErrHandler) func(*aggtrade_types.AggTrades) types.StreamFunction {
	return func(at *aggtrade_types.AggTrades) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsAggTradeServe(at.Symbol(), handler(at), errHandler(at))
			return
		}
	}
}

func WsErrorHandlerCreator() func(*aggtrade_types.AggTrades) binance.ErrHandler {
	return func(trade *aggtrade_types.AggTrades) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Spot AggTrades error: %v", err)
			trade.ResetEvent(err)
		}
	}
}
