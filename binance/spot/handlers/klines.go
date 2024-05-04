package handlers

import (
	"github.com/adshao/go-binance/v2"
	kline_types "github.com/fr0ster/go-trading-utils/types/kline"
)

func GetKlinesUpdateGuard(klines *kline_types.Klines, source chan *binance.WsKlineEvent) (out chan bool) {
	out = make(chan bool)
	go func() {
		for {
			event := <-source
			if klines.Time >= event.Time {
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
			}
			klines.Lock() // Locking the bookTickers
			klines.SetKline(kline)
			klines.Unlock() // Unlocking the bookTickers
			out <- true
		}
	}()
	return
}
