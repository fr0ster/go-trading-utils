package streams

import (
	"github.com/adshao/go-binance/v2/futures"
)

const (
	ChannelUserData   = "userData"
	ChannelDepth      = "depth"
	ChannelKline      = "kline"
	ChannelTrade      = "trade"
	ChannelAggTrade   = "aggTrade"
	ChannelBookTicker = "bookTicker"
)

func StartUserDataStream(listenKey string, out chan *futures.WsUserDataEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsUserDataEvent) {
		out <- event
	}
	return futures.WsUserDataServe(listenKey, wsHandler, handleErr)
}

func StartPartialDepthStream(symbol string, levels int, out chan *futures.WsDepthEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		out <- event
	}
	return futures.WsPartialDepthServe(symbol, levels, wsHandler, handleErr)
}

func StartDiffDepthStream(symbol string, out chan *futures.WsDepthEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		out <- event
	}
	return futures.WsDiffDepthServe(symbol, wsHandler, handleErr)
}

func StartCombinedDepthStream(symbols map[string]string, out chan *futures.WsDepthEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsDepthEvent) {
		out <- event
	}
	return futures.WsCombinedDepthServe(symbols, wsHandler, handleErr)
}

func StartKlineStream(symbol string, interval string, out chan *futures.WsKlineEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsKlineEvent) {
		out <- event
	}
	return futures.WsKlineServe(symbol, interval, wsHandler, handleErr)
}

func StartCombinedAggTradeStream(symbols []string, out chan *futures.WsAggTradeEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsAggTradeEvent) {
		out <- event
	}
	return futures.WsCombinedAggTradeServe(symbols, wsHandler, handleErr)
}

func StartAggTradeStream(symbol string, out chan *futures.WsAggTradeEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsAggTradeEvent) {
		out <- event
	}
	return futures.WsAggTradeServe(symbol, wsHandler, handleErr)
}

func StartBookTickerStream(symbol string, out chan *futures.WsBookTickerEvent, handleErr futures.ErrHandler) (doneC, stopC chan struct{}, err error) {
	wsHandler := func(event *futures.WsBookTickerEvent) {
		out <- event
	}
	return futures.WsBookTickerServe(symbol, wsHandler, handleErr)
}
