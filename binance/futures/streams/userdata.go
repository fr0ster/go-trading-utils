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

func NewUserDataStream(listenKey string) *UserDataStream {
	return &UserDataStream{
		DataChannel:  make(chan *futures.WsUserDataEvent),
		EventChannel: make(chan bool),
		listenKey:    listenKey,
	}
}

func (u *UserDataStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsUserDataEvent) {
		u.DataChannel <- event
	}
	return futures.WsUserDataServe(u.listenKey, wsHandler, utils.HandleErr)
}
