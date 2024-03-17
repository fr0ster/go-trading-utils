package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type AggTradeStream struct {
	DataChannel  chan *futures.WsAggTradeEvent
	EventChannel chan bool
	symbol       string
}

func NewTradeStream(symbol string) *AggTradeStream {
	return &AggTradeStream{
		DataChannel:  make(chan *futures.WsAggTradeEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *AggTradeStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *AggTradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsAggTradeEvent) {
		u.DataChannel <- event
	}
	return futures.WsAggTradeServe(u.symbol, wsHandler, utils.HandleErr)
}

type CombinedAggTradeServe struct {
	DataChannel  chan *futures.WsAggTradeEvent
	EventChannel chan bool
	symbols      []string
}

func NewCombinedAggTradeServe(symbols []string) *CombinedAggTradeServe {
	return &CombinedAggTradeServe{
		DataChannel:  make(chan *futures.WsAggTradeEvent),
		EventChannel: make(chan bool),
		symbols:      symbols,
	}
}

func (u *CombinedAggTradeServe) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *CombinedAggTradeServe) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsAggTradeEvent) {
		u.DataChannel <- event
	}
	return futures.WsCombinedAggTradeServe(u.symbols, wsHandler, utils.HandleErr)
}
