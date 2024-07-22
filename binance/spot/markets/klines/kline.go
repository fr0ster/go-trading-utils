package kline

import (
	"context"

	"github.com/adshao/go-binance/v2"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
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
	handler func(*kline_types.Klines) binance.WsKlineHandler,
	errHandler func(*kline_types.Klines) binance.ErrHandler) func(*kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
	return func(kl *kline_types.Klines) func() (doneC, stopC chan struct{}, err error) {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsKlineServe(kl.GetSymbolname(), string(kl.GetInterval()), handler(kl), errHandler(kl))
			return
		}
	}
}

func standardEventHandlerCreator(kl *kline_types.Klines) binance.WsKlineHandler {
	return func(event *binance.WsKlineEvent) {
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
	handlers ...func(*kline_types.Klines) binance.WsKlineHandler) func(*kline_types.Klines) binance.WsKlineHandler {
	return func(kl *kline_types.Klines) binance.WsKlineHandler {
		var stack []binance.WsKlineHandler
		standardHandlers := standardEventHandlerCreator(kl)
		for _, handler := range handlers {
			stack = append(stack, handler(kl))
		}
		return func(event *binance.WsKlineEvent) {
			standardHandlers(event)
			for _, handler := range stack {
				handler(event)
			}
		}
	}
}
