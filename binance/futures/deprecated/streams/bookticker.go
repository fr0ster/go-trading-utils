package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type BookTickerStream struct {
	DataChannel  chan *futures.WsBookTickerEvent
	EventChannel chan bool
	symbol       string
}

func NewBookTickerStream(symbol string, size int) *BookTickerStream {
	return &BookTickerStream{
		DataChannel:  make(chan *futures.WsBookTickerEvent, size),
		EventChannel: make(chan bool, size),
		symbol:       symbol,
	}
}

func (u *BookTickerStream) GetDataChannel() chan *futures.WsBookTickerEvent {
	return u.DataChannel
}

func (u *BookTickerStream) GetEventChannel() chan bool {
	return u.EventChannel
}

func (u *BookTickerStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsBookTickerEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsBookTickerServe(u.symbol, wsHandler, utils.HandleErr)
}
