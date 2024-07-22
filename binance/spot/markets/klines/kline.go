package kline

import (
	"context"

	"github.com/adshao/go-binance/v2"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
	"github.com/fr0ster/go-trading-utils/types/ring_buffer"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

func GetInitCreator(interval kline_types.KlineStreamInterval, client *binance.Client) func(*kline_types.Klines) func() (err error) {
	return func(kl *kline_types.Klines) func() (err error) {
		return func() (err error) {
			kl.Lock()         // Locking the klines
			defer kl.Unlock() // Unlocking the klines
			klines, _ :=
				client.NewKlinesService().
					Interval(string(interval)).
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
	}
}

func GetStartKlineStreamCreator(
	handler binance.WsKlineHandler,
	errHandler binance.ErrHandler) func(d *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
	return func(pp *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsKlineServe(pp.GetSymbolname(), string(pp.GetInterval()), handler, errHandler)
			return
		}
	}
}

func GetKlineCallBackCreator(
	maxRing,
	minRing *ring_buffer.RingBuffer) func(kl *kline_types.Klines) binance.WsKlineHandler {
	return func(kl *kline_types.Klines) binance.WsKlineHandler {
		return func(event *binance.WsKlineEvent) {
			high := utils.ConvStrToFloat64(event.Kline.High)
			low := utils.ConvStrToFloat64(event.Kline.Low)
			if event.Kline.IsFinal {
				maxRing.Add(high)
				minRing.Add(low)
			}
		}
	}
}
