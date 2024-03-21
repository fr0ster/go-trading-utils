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

func NewBookTickerStream(symbol string) *BookTickerStream {
	return &BookTickerStream{
		DataChannel:  make(chan *futures.WsBookTickerEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *BookTickerStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *BookTickerStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsBookTickerEvent) {
		u.DataChannel <- event
		u.EventChannel <- true
	}
	return futures.WsBookTickerServe(u.symbol, wsHandler, utils.HandleErr)
}
