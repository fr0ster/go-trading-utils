package kline

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
	"github.com/fr0ster/go-trading-utils/types/ring_buffer"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func GetInitCreator(client *futures.Client) func(kl *kline_types.Klines) func() (err error) {
	return func(kl *kline_types.Klines) func() (err error) {
		return func() (err error) {
			kl.Lock()         // Locking the klines
			defer kl.Unlock() // Unlocking the klines
			klines, _ :=
				client.NewKlinesService().
					Interval(string(kl.GetInterval())).
					Symbol(kl.GetSymbolname()).
					Do(context.Background())
			for _, kline := range klines {
				klineItem, err := kline_types.Binance2kline(kline)
				if err != nil {
					return err
				}
				kl.SetKline(klineItem)
			}
			logrus.Debugf("Futures, Klines size for %v - %v Klines", kl.GetSymbolname(), kl.GetKlines().Len())
			return nil
		}
	}
}

func GetStartKlineStreamCreator(
	handler futures.WsKlineHandler,
	errHandler futures.ErrHandler) func(d *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
	return func(pp *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsKlineServe(pp.GetSymbolname(), string(pp.GetInterval()), handler, errHandler)
			return
		}
	}
}

func GetKlineCallBackCreator(
	maxRing,
	minRing *ring_buffer.RingBuffer) func(kl *kline_types.Klines) futures.WsKlineHandler {
	return func(kl *kline_types.Klines) futures.WsKlineHandler {
		return func(event *futures.WsKlineEvent) {
			high := utils.ConvStrToFloat64(event.Kline.High)
			low := utils.ConvStrToFloat64(event.Kline.Low)
			if event.Kline.IsFinal {
				maxRing.Add(high)
				minRing.Add(low)
			}
		}
	}
}
