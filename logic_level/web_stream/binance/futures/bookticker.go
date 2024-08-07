package futures_api

import (
	web_stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
)

func NewBookTickersStream(symbol string, useTestNet bool, websocketKeepalive ...bool) *web_stream.BookTickersStream {
	return web_stream.NewBookTickersStream(symbol, useTestNet, GetWsBaseUrl(useTestNet), websocketKeepalive...)
}
