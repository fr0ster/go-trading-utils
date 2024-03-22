package kline

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func Init(d *kline_types.Kline, apt_key string, secret_key string, symbolname string, UseTestnet bool) {
	futures.UseTestnet = UseTestnet
	klines, _ :=
		futures.NewClient(apt_key, secret_key).NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		d.Set(kline_types.KlineItem(*kline))
	}
}
