package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type CombinedMarketStatStream struct {
	DataChannel  chan *binance.WsMarketStatEvent
	EventChannel chan bool
	symbols      []string
}

func NewCombinedMarketStatStream(symbols []string) *CombinedMarketStatStream {
	return &CombinedMarketStatStream{
		DataChannel:  make(chan *binance.WsMarketStatEvent),
		EventChannel: make(chan bool),
		symbols:      symbols,
	}
}

func (u *CombinedMarketStatStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *CombinedMarketStatStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsMarketStatEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return binance.WsCombinedMarketStatServe(u.symbols, wsHandler, utils.HandleErr)
}

type AllMiniMarketsStaStream struct {
	DataChannel  chan binance.WsAllMiniMarketsStatEvent
	EventChannel chan bool
	symbols      []string
}

func NewAllMiniMarketsStaStream(symbols []string) *AllMiniMarketsStaStream {
	return &AllMiniMarketsStaStream{
		DataChannel:  make(chan binance.WsAllMiniMarketsStatEvent),
		EventChannel: make(chan bool),
		symbols:      symbols,
	}
}

func (u *AllMiniMarketsStaStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *AllMiniMarketsStaStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event binance.WsAllMiniMarketsStatEvent) {
		u.DataChannel <- event
		u.EventChannel <- true
	}
	return binance.WsAllMiniMarketsStatServe(wsHandler, utils.HandleErr)
}
