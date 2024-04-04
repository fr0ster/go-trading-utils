package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type BookTickerStream struct {
	DataChannel  chan *binance.WsBookTickerEvent
	EventChannel chan bool
	symbol       string
}

func NewBookTickerStream(symbol string, size int) *BookTickerStream {
	return &BookTickerStream{
		DataChannel:  make(chan *binance.WsBookTickerEvent, size),
		EventChannel: make(chan bool, size),
		symbol:       symbol,
	}
}

func (u *BookTickerStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *BookTickerStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsBookTickerEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return binance.WsBookTickerServe(u.symbol, wsHandler, utils.HandleErr)
}
