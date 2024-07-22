package trade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/trade"
)

func tradesInit(trd []*futures.Trade, a *trade_types.Trades) (err error) {
	for _, val := range trd {
		a.Update(&trade_types.Trade{
			ID:           val.ID,
			Price:        val.Price,
			Quantity:     val.Quantity,
			Time:         val.Time,
			IsBuyerMaker: val.IsBuyerMaker,
			// IsBestMatch:  val.IsBestMatch,
			// IsIsolated:   val.IsIsolated,
		})
	}
	return nil
}

func GetHistoricalTradesInitCreator(client *futures.Client, limit int) func(a *trade_types.Trades) func() (err error) {
	return func(a *trade_types.Trades) func() (err error) {
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

func GetRecentTradesInitCreator(client *futures.Client, limit int) func(a *trade_types.Trades) func() (err error) {
	return func(a *trade_types.Trades) func() (err error) {
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
