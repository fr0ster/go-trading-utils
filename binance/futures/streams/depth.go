package streams

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type PartialDepthStream struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	levels       int
	symbol       string
}

func NewDepthStream(symbol string, levels int) *PartialDepthStream {
	return &PartialDepthStream{
		DataChannel:  make(chan *futures.WsDepthEvent),
		EventChannel: make(chan bool),
		levels:       levels,
		symbol:       symbol,
	}
}

func (u *PartialDepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		u.DataChannel <- event
	}
	return futures.WsPartialDepthServe(u.symbol, u.levels, wsHandler, utils.HandleErr)
}

type PartialDepthServeWithRate struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	levels       int
	rate         time.Duration
	symbol       string
}

func NewDepthStreamWithRate(symbol string, levels int, rate time.Duration) *PartialDepthServeWithRate {
	return &PartialDepthServeWithRate{
		DataChannel:  make(chan *futures.WsDepthEvent),
		EventChannel: make(chan bool),
		levels:       levels,
		symbol:       symbol,
		rate:         rate,
	}
}

func (u *PartialDepthServeWithRate) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		u.DataChannel <- event
	}
	return futures.WsPartialDepthServeWithRate(u.symbol, u.levels, u.rate, wsHandler, utils.HandleErr)
}

type DiffDepthServe struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	symbol       string
}

func NewDiffDepthStream(symbol string) *DiffDepthServe {
	return &DiffDepthServe{
		DataChannel:  make(chan *futures.WsDepthEvent),
		EventChannel: make(chan bool),
		symbol:       symbol,
	}
}

func (u *DiffDepthServe) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		u.DataChannel <- event
	}
	return futures.WsDiffDepthServe(u.symbol, wsHandler, utils.HandleErr)
}

type CombinedDepthServe struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	symbolLevels map[string]string
}

func NewCombinedDepthStream(symbolLevels map[string]string) *CombinedDepthServe {
	return &CombinedDepthServe{
		DataChannel:  make(chan *futures.WsDepthEvent),
		EventChannel: make(chan bool),
		symbolLevels: symbolLevels,
	}
}

func (u *CombinedDepthServe) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		u.DataChannel <- event
	}
	return futures.WsCombinedDepthServe(u.symbolLevels, wsHandler, utils.HandleErr)
}
