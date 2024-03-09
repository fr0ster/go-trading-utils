package streams

import (
	"log"

	"github.com/adshao/go-binance/v2"
)

func StartUserDataStream(listenKey string, wsHandler binance.WsUserDataHandler, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error, streamChan chan binance.WsUserDataEvent) {
	streamChan = make(chan binance.WsUserDataEvent)
	doneC, stopC, err = binance.WsUserDataServe(listenKey, wsHandler, handleErr)
	if err != nil {
		log.Fatalf("Error serving user data websocket: %v", err)
		streamChan = nil
		return
	}
	return
}
