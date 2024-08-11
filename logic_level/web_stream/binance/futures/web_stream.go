package futures_web_stream

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
	stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common/stream"

	web_api "github.com/fr0ster/turbo-restler/web_api"
)

type WebStream interface {
	Klines(interval string) *stream.Stream
	Depths(level common.DepthStreamLevel) *stream.Stream
	BookTickers() *stream.Stream
	MiniTickers() *stream.Stream
	AggTrades() *stream.Stream
	UserData(listenKey string) *stream.Stream
}

func New(symbol string, useTestNet ...bool) WebStream {
	var (
		wsEndpoint string
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	if useTestNet[0] {
		wsEndpoint = "testnet.binancefuture.com"
	} else {
		wsEndpoint = "ws-fapi.binance.com"
	}
	return common.New(web_api.WsHost(wsEndpoint), symbol)
}
