package kline

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
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
	handler func(*kline_types.Klines) futures.WsKlineHandler,
	errHandler func(*kline_types.Klines) futures.ErrHandler) func(*kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
	return func(kl *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = futures.WsKlineServe(kl.GetSymbolname(), string(kl.GetInterval()), handler(kl), errHandler(kl))
			return
		}
	}
}

func standardEventHandlerCreator(kl *kline_types.Klines) futures.WsKlineHandler {
	return func(event *futures.WsKlineEvent) {
		func() {
			kl.Lock()         // Locking the depths
			defer kl.Unlock() // Unlocking the depths
			if event.Kline.IsFinal {
				klineItem := &kline_types.Kline{
					OpenTime:                 event.Kline.StartTime,
					Open:                     event.Kline.Open,
					High:                     event.Kline.High,
					Low:                      event.Kline.Low,
					Close:                    event.Kline.Close,
					Volume:                   event.Kline.Volume,
					CloseTime:                event.Kline.EndTime,
					QuoteAssetVolume:         event.Kline.QuoteVolume,
					TradeNum:                 event.Kline.TradeNum,
					TakerBuyBaseAssetVolume:  event.Kline.ActiveBuyVolume,
					TakerBuyQuoteAssetVolume: event.Kline.ActiveBuyQuoteVolume,
					IsFinal:                  event.Kline.IsFinal,
				}
				kl.SetKline(klineItem)
			}
		}()
	}
}

func StandardEventCallBackCreator(
	handlers ...func(*kline_types.Klines) futures.WsKlineHandler) func(*kline_types.Klines) futures.WsKlineHandler {
	return func(kl *kline_types.Klines) futures.WsKlineHandler {
		var stack []futures.WsKlineHandler
		standardHandlers := standardEventHandlerCreator(kl)
		for _, handler := range handlers {
			stack = append(stack, handler(kl))
		}
		return func(event *futures.WsKlineEvent) {
			standardHandlers(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}
