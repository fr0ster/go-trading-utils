package spot_web_api

import (
	web_api "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common"
	signature "github.com/fr0ster/go-trading-utils/low_level/utils/signature"
)

func New(apiKey, apiSecret, symbol string, sign signature.Sign, useTestNet ...bool) *web_api.WebApi {
	var (
		waHost  string
		waPath  string
		baseUrl string
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	if useTestNet[0] {
		waHost = "testnet.binance.vision"
		waPath = "/ws-api/v3"
	} else {
		waHost = "ws-api.binance.com:443"
		waPath = "/ws-api/v3"
	}
	baseUrl = GetWsBaseUrl(useTestNet...)
	return web_api.NewWebApi(apiKey, apiSecret, symbol, baseUrl, waHost, waPath, sign)
}
