package streams

// import (
// 	"github.com/adshao/go-binance/v2"
// )

// const (
// 	ChannelUserData   = "userData"
// 	ChannelDepth      = "depth"
// 	ChannelKline      = "kline"
// 	ChannelTrade      = "trade"
// 	ChannelAggTrade   = "aggTrade"
// 	ChannelBookTicker = "bookTicker"
// )

// func StartUserDataStream(listenKey string, out chan *binance.WsUserDataEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsUserDataEvent) {
// 		out <- event
// 	}
// 	return binance.WsUserDataServe(listenKey, wsHandler, handleErr)
// }

// func StartDepthStream(symbol string, out chan *binance.WsDepthEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsDepthEvent) {
// 		out <- event
// 	}
// 	return binance.WsDepthServe(symbol, wsHandler, handleErr)
// }

// func StartKlineStream(symbol string, interval string, out chan *binance.WsKlineEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsKlineEvent) {
// 		out <- event
// 	}
// 	return binance.WsKlineServe(symbol, interval, wsHandler, handleErr)
// }

// func StartTradeStream(symbol string, out chan *binance.WsTradeEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsTradeEvent) {
// 		out <- event
// 	}
// 	return binance.WsTradeServe(symbol, wsHandler, handleErr)
// }

// func StartAggTradeStream(symbol string, out chan *binance.WsAggTradeEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsAggTradeEvent) {
// 		out <- event
// 	}
// 	return binance.WsAggTradeServe(symbol, wsHandler, handleErr)
// }

// func StartBookTickerStream(symbol string, out chan *binance.WsBookTickerEvent, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
// 	wsHandler := func(event *binance.WsBookTickerEvent) {
// 		out <- event
// 	}
// 	return binance.WsBookTickerServe(symbol, wsHandler, handleErr)
// }
