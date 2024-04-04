package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type DepthStream struct {
	DataChannel  chan *binance.WsDepthEvent
	EventChannel chan bool
	symbol       string
	use100Ms     bool
}

func NewDepthStream(symbol string, use100Ms bool, size int) *DepthStream {
	return &DepthStream{
		DataChannel:  make(chan *binance.WsDepthEvent, size),
		EventChannel: make(chan bool, size),
		symbol:       symbol,
		use100Ms:     use100Ms,
	}
}

func (u *DepthStream) GetStreamEvent() chan bool {
	return u.EventChannel
}

func (u *DepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsDepthEvent) {
		go func() {
			u.DataChannel <- event
		}()
		go func() {
			u.EventChannel <- true
		}()
	}
	if u.use100Ms {
		return binance.WsDepthServe100Ms(u.symbol, wsHandler, utils.HandleErr)
	} else {
		return binance.WsDepthServe(u.symbol, wsHandler, utils.HandleErr)
	}
}
