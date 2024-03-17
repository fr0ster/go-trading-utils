package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type AggTradeStream struct {
	DataChannel  chan *binance.WsAggTradeEvent
	EventChannel chan bool
	symbol       string
}

func NewAggTradeStream(symbol string) *AggTradeStream {
	return &AggTradeStream{
		DataChannel:  make(chan *binance.WsAggTradeEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *AggTradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsAggTradeEvent) {
		u.DataChannel <- event
	}
	return binance.WsAggTradeServe(u.symbol, wsHandler, utils.HandleErr)
}
