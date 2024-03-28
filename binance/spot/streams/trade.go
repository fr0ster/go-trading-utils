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

func (u *TradeStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *TradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsTradeEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return binance.WsTradeServe(u.symbol, wsHandler, utils.HandleErr)
}

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

func (u *AggTradeStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *AggTradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsAggTradeEvent) {
		u.DataChannel <- event
	}
	return binance.WsAggTradeServe(u.symbol, wsHandler, utils.HandleErr)
}

type CombinedTradeStream struct {
	DataChannel  chan *binance.WsCombinedTradeEvent
	EventChannel chan bool
	symbols      []string
}

func NewCombinedTradeStream(symbols []string) *CombinedTradeStream {
	return &CombinedTradeStream{
		DataChannel:  make(chan *binance.WsCombinedTradeEvent),
		EventChannel: make(chan bool),
		symbols:      symbols,
	}
}

func (u *CombinedTradeStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *CombinedTradeStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsCombinedTradeEvent) {
		u.DataChannel <- event
		u.EventChannel <- true
	}
	return binance.WsCombinedTradeServe(u.symbols, wsHandler, utils.HandleErr)
}
