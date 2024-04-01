package trade

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func AggTradeInit(at *trade_types.AggTrades, apt_key, secret_key, symbolname string, limit int, UseTestnet bool) (err error) {
	at.Lock()         // Locking the aggTrades
	defer at.Unlock() // Unlocking the aggTrades
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(apt_key, secret_key)
	res, err :=
		client.NewAggTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, trade := range res {
		at.Update(&trade_types.AggTrade{
			AggTradeID:   trade.AggTradeID,
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			Timestamp:    trade.Timestamp,
			IsBuyerMaker: trade.IsBuyerMaker,
		})
	}
	return nil
}
