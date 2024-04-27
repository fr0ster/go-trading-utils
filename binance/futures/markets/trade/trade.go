package trade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func tradesInit(trd []*futures.Trade, a *trade_types.Trades) (err error) {
	a.Lock()         // Locking the trades
	defer a.Unlock() // Unlocking the trades
	for _, val := range trd {
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

func HistoricalTradesInit(a *trade_types.Trades, client *futures.Client, symbolname string, limit int) (err error) {
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

func RecentTradesInit(a *trade_types.Trades, client *futures.Client, symbolname string, limit int) (err error) {
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
