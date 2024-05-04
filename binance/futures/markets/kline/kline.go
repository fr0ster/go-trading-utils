package kline

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	"github.com/sirupsen/logrus"
)

func Init(kl *kline_types.Klines, client *futures.Client, symbolname string) (err error) {
	kl.Lock()         // Locking the klines
	defer kl.Unlock() // Unlocking the klines
	klines, _ :=
		client.NewKlinesService().
			Interval(kl.GetInterval()).
			Symbol(symbolname).
			Do(context.Background())
	for _, kline := range klines {
		klineItem, err := kline_types.Binance2kline(kline)
		if err != nil {
			return err
		}
		kl.SetKline(klineItem)
	}
	logrus.Debugf("Futures, Klines size for %v - %v Klines", symbolname, kl.GetKlines().Len())
	return nil
}
