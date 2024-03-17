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

func NewBookTickerStream(symbol string) *BookTickerStream {
	return &BookTickerStream{
		DataChannel:  make(chan *binance.WsBookTickerEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *BookTickerStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsBookTickerEvent) {
		u.DataChannel <- event
	}
	return binance.WsBookTickerServe(u.symbol, wsHandler, utils.HandleErr)
}
