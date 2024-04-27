package kline

import (
	"context"

	"github.com/adshao/go-binance/v2"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func Init(kl *kline_types.Kline, client *binance.Client, symbolname string) (err error) {
	kl.Lock()         // Locking the klines
	defer kl.Unlock() // Unlocking the klines
	klines, _ :=
		client.NewKlinesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	for _, kline := range klines {
		klineItem, err := kline_types.Binance2kline(kline)
		if err != nil {
			return err
		}
		kl.Set(klineItem)
	}
	return nil
}
