package kline

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	kline_types "github.com/fr0ster/go-trading-utils/types/klines"
	"github.com/sirupsen/logrus"
)

func InitCreator(client *binance.Client) func(*kline_types.Klines) types.InitFunction {
	return func(kl *kline_types.Klines) types.InitFunction {
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
			logrus.Debugf("Spot, Klines size for %v - %v klines", kl.GetSymbolname(), kl.GetKlines().Len())
			return nil
		}
	}
}

func KlineStreamCreator(
	handler func(*kline_types.Klines) binance.WsKlineHandler,
	errHandler func(*kline_types.Klines) binance.ErrHandler) func(*kline_types.Klines) types.StreamFunction {
	return func(kl *kline_types.Klines) types.StreamFunction {
		return func() (doneC, stopC chan struct{}, err error) {
			// Запускаємо стрім подій користувача
			doneC, stopC, err = binance.WsKlineServe(kl.GetSymbolname(), string(kl.GetInterval()), handler(kl), errHandler(kl))
			return
		}
	}
}

func eventHandlerCreator(kl *kline_types.Klines) binance.WsKlineHandler {
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

func CallBackCreator(
	handlers ...func(*kline_types.Klines) binance.WsKlineHandler) func(*kline_types.Klines) binance.WsKlineHandler {
	return func(kl *kline_types.Klines) binance.WsKlineHandler {
		var stack []binance.WsKlineHandler
		standardHandlers := eventHandlerCreator(kl)
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

func WsErrorHandlerCreator() func(*kline_types.Klines) binance.ErrHandler {
	return func(kl *kline_types.Klines) binance.ErrHandler {
		return func(err error) {
			logrus.Errorf("Spot wsErrorHandler error: %v", err)
			kl.ResetEvent(err)
		}
	}
}
