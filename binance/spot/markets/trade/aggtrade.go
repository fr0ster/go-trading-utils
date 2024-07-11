package trade

import (
	"context"

	"github.com/adshao/go-binance/v2"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
)

func AggTradeInit(at *trade_types.AggTrades, client *binance.Client, symbolname string, limit int) (err error) {
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
