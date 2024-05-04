package handlers

import (
	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
	"github.com/sirupsen/logrus"
)

func GetKlinesUpdateGuard(klines *kline_types.Klines, source chan *futures.WsKlineEvent, IsFinal bool) (
	finalOut chan bool, nonFinalOut chan bool) {
	finalOut = make(chan bool, 1)
	nonFinalOut = make(chan bool, 1)
	logrus.Debug("Create Kline updater goroutine")
	go func() {
		for {
			event := <-source
			if IsFinal && !event.Kline.IsFinal {
				continue
			}
			if klines.GetTime() >= event.Time {
				continue
			}
			kline := &kline_types.Kline{
				OpenTime:                 event.Kline.StartTime,
				CloseTime:                event.Kline.EndTime,
				Open:                     event.Kline.Open,
				Close:                    event.Kline.Close,
				High:                     event.Kline.High,
				Low:                      event.Kline.Low,
				Volume:                   event.Kline.Volume,
				QuoteAssetVolume:         event.Kline.QuoteVolume,
				TradeNum:                 event.Kline.TradeNum,
				TakerBuyBaseAssetVolume:  event.Kline.ActiveBuyVolume,
				TakerBuyQuoteAssetVolume: event.Kline.ActiveBuyQuoteVolume,
				IsFinal:                  event.Kline.IsFinal,
			}
			klines.Lock() // Locking the klines
			klines.SetTime(event.Time)
			if event.Kline.IsFinal {
				klines.SetKline(kline)
				finalOut <- true
			} else {
				klines.SetLastKline(kline)
				nonFinalOut <- true
			}
			klines.Unlock() // Unlocking the klines
		}
	}()
	return
}
