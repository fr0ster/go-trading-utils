package streams

import (
	"github.com/adshao/go-binance/v2"
)

func StartUserDataStream(listenKey string, wsHandler binance.WsUserDataHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsUserDataServe(listenKey, wsHandler, handleErr)
}

func StartDepthStream(symbol string, wsHandler binance.WsDepthHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsDepthServe(symbol, wsHandler, handleErr)
}

func StartKlineStream(symbol string, interval string, wsHandler binance.WsKlineHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsKlineServe(symbol, interval, wsHandler, handleErr)
}

func StartTradeStream(symbol string, wsHandler binance.WsTradeHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsTradeServe(symbol, wsHandler, handleErr)
}

func StartAggTradeStream(symbol string, wsHandler binance.WsAggTradeHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsAggTradeServe(symbol, wsHandler, handleErr)
}

func StartBookTickerStream(symbol string, wsHandler binance.WsBookTickerHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	return binance.WsBookTickerServe(symbol, wsHandler, handleErr)
}
