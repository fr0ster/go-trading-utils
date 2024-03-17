package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type KlineStream struct {
	DataChannel  chan *futures.WsKlineEvent
	EventChannel chan bool
	interval     string
	symbol       string
}

func NewKlineStream(symbol, interval string) *KlineStream {
	return &KlineStream{
		DataChannel:  make(chan *futures.WsKlineEvent),
		EventChannel: make(chan bool),
		interval:     interval,
		symbol:       symbol,
	}
}

func (u *KlineStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *KlineStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsKlineEvent) {
		u.DataChannel <- event
	}
	return futures.WsKlineServe(u.symbol, u.interval, wsHandler, utils.HandleErr)
}
