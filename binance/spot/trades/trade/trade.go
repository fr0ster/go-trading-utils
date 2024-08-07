package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/trade"
	"github.com/sirupsen/logrus"
)

func tradesInit(trd []*binance.Trade, a *trade_types.Trades) (err error) {
	for _, val := range trd {
		a.Update(&trade_types.Trade{
			ID:           val.ID,
			Price:        val.Price,
			Quantity:     val.Quantity,
			Time:         val.Time,
			IsBuyerMaker: val.IsBuyerMaker,
			IsBestMatch:  val.IsBestMatch,
			IsIsolated:   val.IsIsolated,
		})
	}
	return nil
}

func HistoricalTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.Trades) types.InitFunction {
	return func(a *trade_types.Trades) types.InitFunction {
		return func() (err error) {
			res, err :=
				client.NewHistoricalTradesService().
					Symbol(string(a.GetSymbolname())).
					Limit(limit).
					Do(context.Background())
			if err != nil {
				return err
			}
			return tradesInit(res, a)
		}
	}
}

func RecentTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.Trades) types.InitFunction {
	return func(a *trade_types.Trades) types.InitFunction {
		return func() (err error) {
			res, err :=
				client.NewRecentTradesService().
					Symbol(string(a.GetSymbolname())).
					Limit(limit).
					Do(context.Background())
			if err != nil {
				return err
			}
			return tradesInit(res, a)
		}
	}
}

func WsErrorHandlerCreator() func(*trade_types.Trades) binance.ErrHandler {
	return func(trade *trade_types.Trades) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Spot Trades error: %v", err)
			trade.ResetEvent(err)
		}
	}
}
