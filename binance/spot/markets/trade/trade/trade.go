package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade/trade"
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

func GetHistoricalTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.Trades) func() (err error) {
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

func GetRecentTradesInitCreator(client *binance.Client, limit int) func(a *trade_types.Trades) func() (err error) {
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
