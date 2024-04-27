package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func AggTradeInit(at *trade_types.AggTrades, client *binance.Client, symbolname string, limit int) (err error) {
	at.Lock()         // Locking the aggTrades
	defer at.Unlock() // Unlocking the aggTrades
	res, err :=
		client.NewAggTradesService().
			Symbol(string(symbolname)).
			Limit(limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, val := range res {
		aggTrade, err := trade_types.Binance2AggTrades(val)
		if err != nil {
			return err
		}
		at.Update(aggTrade)
	}
	return nil
}
