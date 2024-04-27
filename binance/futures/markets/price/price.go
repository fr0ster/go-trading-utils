package price

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
)

func Init24h(prc *price_types.PriceChangeStats, client futures.Client, symbols ...string) (err error) {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	var pcss []*futures.PriceChangeStats
	if len(symbols) > 0 {
		for _, symbol := range symbols {
			res, _ :=
				client.NewListPriceChangeStatsService().Symbol(symbol).Do(context.Background())
			pcss = append(pcss, res...)
		}
	} else {
		pcss, err =
			client.NewListPriceChangeStatsService().Do(context.Background())
		if err != nil {
			return err
		}
	}
	for _, pcs := range pcss {
		prc.Set(&price_types.PriceChangeStat{
			Symbol:             pcs.Symbol,
			PriceChange:        pcs.PriceChange,
			PriceChangePercent: pcs.PriceChangePercent,
			WeightedAvgPrice:   pcs.WeightedAvgPrice,
			PrevClosePrice:     pcs.PrevClosePrice,
			LastPrice:          pcs.LastPrice,
			LastQty:            pcs.LastQuantity,
			OpenPrice:          pcs.OpenPrice,
			HighPrice:          pcs.HighPrice,
			LowPrice:           pcs.LowPrice,
			Volume:             pcs.Volume,
			QuoteVolume:        pcs.QuoteVolume,
			OpenTime:           pcs.OpenTime,
			CloseTime:          pcs.CloseTime,
			FirstID:            pcs.FirstID,
			LastID:             pcs.LastID,
			Count:              pcs.Count,
		})
	}
	return nil
}
