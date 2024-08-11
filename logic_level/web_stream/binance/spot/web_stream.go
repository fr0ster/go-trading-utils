package futures_web_stream

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
	stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common/stream"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

type WebStream interface {
	Klines(interval string) *stream.Stream
	Depths(level common.DepthStreamLevel) *stream.Stream
	BookTickers() *stream.Stream
	Trades() *stream.Stream
	MiniTickers() *stream.Stream
	UserData(listenKey string) *stream.Stream
}

func New(symbol string, sign signature.Sign, useTestNet ...bool) WebStream {
	var (
		wsEndpoint string
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	if useTestNet[0] {
		wsEndpoint = "wss://testnet.binance.vision/ws"
	} else {
		wsEndpoint = "wss://stream.binance.com:9443"
	}
	return common.New(symbol, web_api.WsHost(wsEndpoint), symbol, sign)
}
