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
	go func() {
		for {
			event := <-source
			logrus.Debugf("Futures, WsKlineEvent for %v happened", event.Symbol)
			if IsFinal && !event.Kline.IsFinal {
				continue
			}
			logrus.Debugf("Futures, Kline for %v was filled", event.Symbol)
			if klines.GetTime() >= event.Time {
				continue
			}
			logrus.Debugf("Futures, Kline Time for %v was more than event time", event.Symbol)
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
