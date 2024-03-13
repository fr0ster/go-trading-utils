package streams

import (
	"github.com/adshao/go-binance/v2"
)

const (
	ChannelUserData   = "userData"
	ChannelDepth      = "depth"
	ChannelKline      = "kline"
	ChannelTrade      = "trade"
	ChannelAggTrade   = "aggTrade"
	ChannelBookTicker = "bookTicker"
)

func StartUserDataStream(listenKey string, channel chan *binance.WsUserDataEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsUserDataEvent) {
		channel <- event
	}
	return binance.WsUserDataServe(listenKey, wsHandler, handleErr)
}

func StartDepthStream(symbol string, channel chan *binance.WsDepthEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsDepthEvent) {
		channel <- event
	}
	return binance.WsDepthServe(symbol, wsHandler, handleErr)
}

func StartKlineStream(symbol string, interval string, channel chan *binance.WsKlineEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsKlineEvent) {
		channel <- event
	}
	return binance.WsKlineServe(symbol, interval, wsHandler, handleErr)
}

func StartTradeStream(symbol string, channel chan *binance.WsTradeEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsTradeEvent) {
		channel <- event
	}
	return binance.WsTradeServe(symbol, wsHandler, handleErr)
}

func StartAggTradeStream(symbol string, channel chan *binance.WsAggTradeEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsAggTradeEvent) {
		channel <- event
	}
	return binance.WsAggTradeServe(symbol, wsHandler, handleErr)
}

func StartBookTickerStream(symbol string, channel chan *binance.WsBookTickerEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *binance.WsBookTickerEvent) {
		channel <- event
	}
	return binance.WsBookTickerServe(symbol, wsHandler, handleErr)
}
