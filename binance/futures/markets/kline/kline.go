package kline

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func Init(d *kline_types.Kline, apt_key string, secret_key string, symbolname string, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	klines, _ :=
		futures.NewClient(apt_key, secret_key).NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		klineItem, err := kline_types.Binance2kline(kline)
		if err != nil {
			return err
		}
		d.Set(klineItem)
	}
	return nil
}
