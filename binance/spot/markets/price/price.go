package price

import (
	"context"

	"github.com/adshao/go-binance/v2"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
)

func Init24h(prc *price_types.PriceChangeStats, client *binance.Client, symbols ...string) error {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	pcss, _ :=
		client.NewListPriceChangeStatsService().Symbols(symbols).Do(context.Background())
	for _, pcs := range pcss {
		price, err := price_types.Binance2PriceChangeStats(pcs)
		if err != nil {
			return err
		}
		prc.Set(price)
	}
	return nil
}

func Init(prc *price_types.PriceChangeStats, client *binance.Client, symbols ...string) error {
	prc.Lock()         // Locking the price change stats
	defer prc.Unlock() // Unlocking the price change stats
	pcss, _ :=
		client.NewListSymbolTickerService().WindowSize("1m").Symbols(symbols).Do(context.Background())
	for _, pcs := range pcss {
		price, err := price_types.Binance2PriceChangeStats(pcs)
		if err != nil {
			return err
		}
		prc.Set(price)
	}
	return nil
}
