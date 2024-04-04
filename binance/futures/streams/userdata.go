package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type UserDataStream struct {
	DataChannel  chan *futures.WsUserDataEvent
	EventChannel chan bool
	listenKey    string
}

func NewUserDataStream(listenKey string, size int) *UserDataStream {
	return &UserDataStream{
		DataChannel:  make(chan *futures.WsUserDataEvent, size),
		EventChannel: make(chan bool, size),
		listenKey:    listenKey,
	}
}

func (u *UserDataStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *UserDataStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsUserDataEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsUserDataServe(u.listenKey, wsHandler, utils.HandleErr)
}
