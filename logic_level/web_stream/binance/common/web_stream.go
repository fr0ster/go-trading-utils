package common_web_stream

import (
	"fmt"
	"strings"
	"sync"

	streamer "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common/streamer"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

func (wa *WebStream) Klines(interval string) *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@kline_%s", strings.ToLower(wa.symbol), interval))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) Depths(level DepthStreamLevel) *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@depth%s", strings.ToLower(wa.symbol), string(level)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) Depths100ms(level DepthStreamLevel) *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@depth%s@100ms", strings.ToLower(wa.symbol), string(level)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) AggTrades() *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@aggTrade", strings.ToLower(wa.symbol)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) Trades() *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@Trade", strings.ToLower(wa.symbol)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) BookTickers() *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@bookTicker", strings.ToLower(wa.symbol)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) Tickers() *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@ticker", strings.ToLower(wa.symbol)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) MiniTickers() *streamer.Request {
	wsPath := web_api.WsPath(fmt.Sprintf("%s@miniTicker", strings.ToLower(wa.symbol)))
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, wsPath, wa.sign)
}

func (wa *WebStream) UserData(listenKey string) *streamer.Request {
	return streamer.New(wa.apiKey, wa.apiSecret, wa.waHost, web_api.WsPath(listenKey), wa.sign)
}

func New(apiKey, apiSecret string, host web_api.WsHost, symbol string, sign signature.Sign) *WebStream {
	return &WebStream{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		waHost:    host,
		mutex:     &sync.Mutex{},
		sign:      sign,
	}
}
