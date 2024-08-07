package futures_api

import (
	web_stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
)

func NewAggTradeStream(symbol string, useTestNet bool, websocketKeepalive ...bool) *web_stream.AggTradeStream {
	return web_stream.NewAggTradeStream(symbol, useTestNet, GetWsBaseUrl(useTestNet), websocketKeepalive...)
}
