package streams

import (
	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
)

type DepthStream struct {
	dataChannel  chan *binance.WsDepthEvent
	eventChannel chan bool
	symbol       string
	use100Ms     bool
}

func NewDepthStream(symbol string, use100Ms bool, size int) *DepthStream {
	return &DepthStream{
		dataChannel:  make(chan *binance.WsDepthEvent, size),
		eventChannel: make(chan bool, size),
		symbol:       symbol,
		use100Ms:     use100Ms,
	}
}

func (u *DepthStream) GetDataChannel() chan *binance.WsDepthEvent {
	return u.dataChannel
}

func (u *DepthStream) GetEventChannel() chan bool {
	return u.eventChannel
}

func (u *DepthStream) Start() (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsDepthEvent) {
		go func() {
			u.dataChannel <- event
		}()
		go func() {
			u.eventChannel <- true
		}()
	}
	if u.use100Ms {
		return binance.WsDepthServe100Ms(u.symbol, wsHandler, utils.HandleErr)
	} else {
		return binance.WsDepthServe(u.symbol, wsHandler, utils.HandleErr)
	}
}
