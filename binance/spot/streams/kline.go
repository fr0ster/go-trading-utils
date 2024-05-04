package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type KlineStream struct {
	dataChannel  chan *binance.WsKlineEvent
	eventChannel chan bool
	interval     string
	symbol       string
}

func NewKlineStream(symbol, interval string, size int) *KlineStream {
	return &KlineStream{
		dataChannel:  make(chan *binance.WsKlineEvent, size),
		eventChannel: make(chan bool, size),
		interval:     interval,
		symbol:       symbol,
	}
}

func (u *KlineStream) GetDataChannel() chan *binance.WsKlineEvent {
	return u.dataChannel
}

func (u *KlineStream) GetEventChannel() chan bool {
	return u.eventChannel
}

func (u *KlineStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsKlineEvent) {
		go func() {
			u.dataChannel <- event
		}()
		go func() {
			u.eventChannel <- true
		}()
	}
	return binance.WsKlineServe(u.symbol, u.interval, wsHandler, utils.HandleErr)
}
