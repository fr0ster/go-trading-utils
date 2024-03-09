package streams

import (
	"log"

	"github.com/adshao/go-binance/v2"
)

func StartUserDataStream(listenKey string, wsHandler binance.WsUserDataHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	doneC, stopC, err = binance.WsUserDataServe(listenKey, wsHandler, handleErr)
	if err != nil {
		log.Fatalf("Error serving user data websocket: %v", err)
		return
	}
	return
}
