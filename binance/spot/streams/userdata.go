package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type UserDataStream struct {
	DataChannel  chan *binance.WsUserDataEvent
	EventChannel chan bool
	listenKey    string
}

func NewUserDataStream(listenKey string) *UserDataStream {
	return &UserDataStream{
		DataChannel:  make(chan *binance.WsUserDataEvent),
		EventChannel: make(chan bool),
		listenKey:    listenKey,
	}
}

func (u *UserDataStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *UserDataStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsUserDataEvent) {
		u.DataChannel <- event
	}
	return binance.WsUserDataServe(u.listenKey, wsHandler, utils.HandleErr)
}
