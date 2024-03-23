package trade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesInit(res []*futures.Trade, a *trade_types.Trades) (err error) {
	for _, val := range res {
		a.Update(&trade_types.Trade{
			ID:           val.ID,
			Price:        val.Price,
			Quantity:     val.Quantity,
			Time:         val.Time,
			IsBuyerMaker: val.IsBuyerMaker,
		})
	}
	return nil
}

func HistoricalTradesInit(a *trade_types.Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(apt_key, secret_key)
	res, err :=
		client.NewHistoricalTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}

func RecentTradesInit(a *trade_types.Trades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(apt_key, secret_key)
	res, err :=
		client.NewRecentTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	return tradesInit(res, a)
}
