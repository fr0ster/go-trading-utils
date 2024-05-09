package kline

import (
	"context"

	"github.com/adshao/go-binance/v2"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	"github.com/sirupsen/logrus"
)

func Init(kl *kline_types.Klines, client *binance.Client) (err error) {
	kl.Lock()         // Locking the klines
	defer kl.Unlock() // Unlocking the klines
	klines, _ :=
		client.NewKlinesService().
			Interval(kl.GetInterval()).
			Symbol(kl.GetSymbolname()).
			Do(context.Background())
	for _, kline := range klines {
		klineItem, err := kline_types.Binance2kline(kline)
		if err != nil {
			return err
		}
		kl.SetKline(klineItem)
	}
	logrus.Debugf("Spot, Klines size for %v - %v klines", kl.GetSymbolname(), kl.GetKlines().Len())
	return nil
}
