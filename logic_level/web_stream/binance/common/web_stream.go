package common_web_stream

import (
	"fmt"
	"strings"
	"sync"

	stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common/stream"

	web_api "github.com/fr0ster/turbo-restler/web_api"
)

func (wa *WebStream) Klines(interval string) *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@kline_%s", strings.ToLower(wa.symbol), interval))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) Depths(level DepthStreamLevel) *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@depth%s", strings.ToLower(wa.symbol), string(level)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) Depths100ms(level DepthStreamLevel) *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@depth%s@100ms", strings.ToLower(wa.symbol), string(level)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) AggTrades() *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@aggTrade", strings.ToLower(wa.symbol)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) Trades() *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@Trade", strings.ToLower(wa.symbol)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) BookTickers() *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@bookTicker", strings.ToLower(wa.symbol)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) Tickers() *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@ticker", strings.ToLower(wa.symbol)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) MiniTickers() *stream.Stream {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@miniTicker", strings.ToLower(wa.symbol)))
	return stream.New(wa.waHost, wsPath)
}

func (wa *WebStream) UserData(listenKey string) *stream.Stream {
	return stream.New(wa.waHost, web_api.WsPath(listenKey))
}

func New(host web_api.WsHost, symbol string) *WebStream {
	return &WebStream{
		symbol: symbol,
		waHost: host,
		mutex:  &sync.Mutex{},
	}
}
