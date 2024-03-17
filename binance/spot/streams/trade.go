package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type TradeStream struct {
	DataChannel  chan *binance.WsTradeEvent
	EventChannel chan bool
	symbol       string
}

func NewTradeStream(symbol string) *TradeStream {
	return &TradeStream{
		DataChannel:  make(chan *binance.WsTradeEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *TradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsTradeEvent) {
		u.DataChannel <- event
	}
	return binance.WsTradeServe(u.symbol, wsHandler, utils.HandleErr)
}
