package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type DepthStream struct {
	DataChannel  chan *binance.WsDepthEvent
	EventChannel chan bool
	symbol       string
}

func NewDepthStream(symbol string) *DepthStream {
	return &DepthStream{
		DataChannel:  make(chan *binance.WsDepthEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *DepthStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *DepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsDepthEvent) {
		u.DataChannel <- event
	}
	return binance.WsDepthServe(u.symbol, wsHandler, utils.HandleErr)
}
