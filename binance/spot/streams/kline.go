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

func NewKlineStream(symbol, interval string, size int) *KlineStream {
	return &KlineStream{
		DataChannel:  make(chan *binance.WsKlineEvent, size),
		EventChannel: make(chan bool, size),
		interval:     interval,
		symbol:       symbol,
	}
}

func (u *KlineStream) GetEventChannel() chan bool {
	return u.EventChannel
}

func (u *KlineStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsKlineEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return binance.WsKlineServe(u.symbol, u.interval, wsHandler, utils.HandleErr)
}
