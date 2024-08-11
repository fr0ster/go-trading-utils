package common_web_stream

import (
	"fmt"
	"strings"
	"sync"

	streamer "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common/streamer"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

func (wa *WebStream) Klines(interval string) *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s@kline_%s", wa.waHost, strings.ToLower(wa.symbol), interval)
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
}

func (wa *WebStream) Depths() *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s@depth", wa.waHost, strings.ToLower(wa.symbol))
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
}

func (wa *WebStream) AggTrades() *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s@aggTrade", wa.waHost, strings.ToLower(wa.symbol))
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
}

func (wa *WebStream) Trades() *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s@Trade", wa.waHost, strings.ToLower(wa.symbol))
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
}

func (wa *WebStream) BookTickers() *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s@bookTicker", wa.waHost, strings.ToLower(wa.symbol))
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
}

func (wa *WebStream) UserData(listenKey string) *streamer.Streamer {
	wsURL := fmt.Sprintf("%s/%s", wa.waHost, listenKey)
	return streamer.New(wa.apiKey, wa.apiSecret, wsURL, wa.sign)
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
