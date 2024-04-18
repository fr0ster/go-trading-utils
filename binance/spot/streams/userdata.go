package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type UserDataStream struct {
	DataChannel  chan *binance.WsUserDataEvent
	eventChannel chan bool
	listenKey    string
}

func NewUserDataStream(listenKey string, size int) *UserDataStream {
	return &UserDataStream{
		DataChannel:  make(chan *binance.WsUserDataEvent, size),
		eventChannel: make(chan bool, size),
		listenKey:    listenKey,
	}
}

func (u *UserDataStream) GetDataChannel() chan *binance.WsUserDataEvent {
	return u.DataChannel
}

func (u *UserDataStream) GetStreamEvent() chan bool {
	return u.eventChannel
}

func (u *UserDataStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsUserDataEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.eventChannel <- true
		}()
	}
	return binance.WsUserDataServe(u.listenKey, wsHandler, utils.HandleErr)
}
