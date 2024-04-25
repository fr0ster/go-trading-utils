package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type BookTickerStream struct {
	dataChannel  chan *binance.WsBookTickerEvent
	eventChannel chan bool
	symbol       string
}

func NewBookTickerStream(symbol string, size int) *BookTickerStream {
	return &BookTickerStream{
		dataChannel:  make(chan *binance.WsBookTickerEvent, size),
		eventChannel: make(chan bool, size),
		symbol:       symbol,
	}
}

func (u *BookTickerStream) GetDataChannel() chan *binance.WsBookTickerEvent {
	return u.dataChannel
}

func (u *BookTickerStream) GetStreamEvent() chan bool {
	return u.eventChannel
}

func (u *BookTickerStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsBookTickerEvent) {
		go func() {
			u.dataChannel <- event
		}()
		go func() {
			u.eventChannel <- true
		}()
	}
	return binance.WsBookTickerServe(u.symbol, wsHandler, utils.HandleErr)
}
