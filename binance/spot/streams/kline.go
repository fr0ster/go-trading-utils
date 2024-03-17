package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type KlineStream struct {
	DataChannel  chan *binance.WsKlineEvent
	EventChannel chan bool
	interval     string
	symbol       string
}

func NewKlineStream(symbol, interval string) *KlineStream {
	return &KlineStream{
		DataChannel:  make(chan *binance.WsKlineEvent),
		EventChannel: make(chan bool),
		interval:     interval,
		symbol:       symbol,
	}
}

func (u *KlineStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsKlineEvent) {
		u.DataChannel <- event
	}
	return binance.WsKlineServe(u.symbol, u.interval, wsHandler, utils.HandleErr)
}
