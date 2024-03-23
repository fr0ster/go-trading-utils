package price

import (
	"context"

	"github.com/adshao/go-binance/v2"
	price_types "github.com/fr0ster/go-trading-utils/types/price"
)

func Init(d *price_types.PriceChangeStats, apt_key string, secret_key string, symbolname string, UseTestnet bool) error {
	binance.UseTestnet = UseTestnet
	pcss, _ :=
		binance.NewClient(apt_key, secret_key).
			NewListPriceChangeStatsService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, pcs := range pcss {
		price, err := price_types.Binance2PriceChangeStats(pcs)
		if err != nil {
			return err
		}
		d.Set(price)
	}
	return nil
}
