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
