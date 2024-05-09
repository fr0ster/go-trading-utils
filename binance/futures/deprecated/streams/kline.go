package streams

import (
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"
)

type KlineStream struct {
	interval string
	symbol   string
}

func NewKlineStream(symbol, interval string, size int) *KlineStream {
	return &KlineStream{
		interval: interval,
		symbol:   symbol,
	}
}

func (u *KlineStream) Start(handler futures.WsKlineHandler) (doneC, stopC chan struct{}, err error) {
	logrus.Debugf("Futures, Start stream for %v Klines", u.symbol)
	return futures.WsKlineServe(u.symbol, u.interval, handler, utils.HandleErr)
}
