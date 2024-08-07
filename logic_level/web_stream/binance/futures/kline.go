package futures_api

import (
	web_stream "github.com/fr0ster/go-trading-utils/logic_level/web_stream/binance/common"
)

func NewKlineStream(symbol string, useTestNet bool, websocketKeepalive ...bool) *web_stream.KlinesStream {
	return web_stream.NewKlineStream(symbol, useTestNet, GetWsBaseUrl(useTestNet), websocketKeepalive...)
}
