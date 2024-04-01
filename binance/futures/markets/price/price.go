package price

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
)

func Init(prc *price_types.PriceChangeStats, apt_key string, secret_key string, symbolname string, UseTestnet bool) error {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	binance.UseTestnet = UseTestnet
	pcss, err :=
		futures.NewClient(apt_key, secret_key).
			NewListPriceChangeStatsService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, pcs := range pcss {
		prc.Set(&price_types.PriceChangeStatsItem{
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
