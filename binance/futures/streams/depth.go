package streams

import (
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
)

type (
	PartialDepthStream struct {
		DataChannel  chan *futures.WsDepthEvent
		EventChannel chan bool
		levels       int
		symbol       string
	}
)

func NewPartialDepthStream(symbol string, levels int, size int) *PartialDepthStream {
	return &PartialDepthStream{
		DataChannel:  make(chan *futures.WsDepthEvent, size),
		EventChannel: make(chan bool, size),
		levels:       levels,
		symbol:       symbol,
	}
}

func (u *PartialDepthStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *PartialDepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsPartialDepthServe(u.symbol, u.levels, wsHandler, utils.HandleErr)
}

type Rate time.Duration

const (
	Rate100Ms Rate = Rate(100 * time.Millisecond)
	Rate250Ms Rate = Rate(250 * time.Millisecond)
	Rate500Ms Rate = Rate(500 * time.Millisecond)
)

type PartialDepthServeWithRate struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	levels       int
	rate         time.Duration
	symbol       string
}

func NewPartialDepthStreamWithRate(symbol string, levels int, rate Rate, size int) *PartialDepthServeWithRate {
	return &PartialDepthServeWithRate{
		DataChannel:  make(chan *futures.WsDepthEvent, size),
		EventChannel: make(chan bool, size),
		levels:       levels,
		symbol:       symbol,
		rate:         time.Duration(rate),
	}
}

func (u *PartialDepthServeWithRate) GetDataChannel() chan *futures.WsDepthEvent {
	return u.DataChannel
}

func (u *PartialDepthServeWithRate) GetEventChannel() chan bool {
	return u.EventChannel
}

func (u *PartialDepthServeWithRate) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsPartialDepthServeWithRate(u.symbol, u.levels, u.rate, wsHandler, utils.HandleErr)
}

type DiffDepthStream struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	symbol       string
}

func NewDiffDepthStream(symbol string, size int) *DiffDepthStream {
	return &DiffDepthStream{
		DataChannel:  make(chan *futures.WsDepthEvent, size),
		EventChannel: make(chan bool, size),
		symbol:       symbol,
	}
}

func (u *DiffDepthStream) GetEventChannel() chan bool {
	return u.EventChannel
}

func (u *DiffDepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsDiffDepthServe(u.symbol, wsHandler, utils.HandleErr)
}

type CombinedDepthStream struct {
	DataChannel  chan *futures.WsDepthEvent
	EventChannel chan bool
	symbolLevels map[string]string
}

func NewCombinedDepthStream(symbolLevels map[string]string, size int) *CombinedDepthStream {
	return &CombinedDepthStream{
		DataChannel:  make(chan *futures.WsDepthEvent, size),
		EventChannel: make(chan bool, size),
		symbolLevels: symbolLevels,
	}
}

func (u *CombinedDepthStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *CombinedDepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	return futures.WsCombinedDepthServe(u.symbolLevels, wsHandler, utils.HandleErr)
}
