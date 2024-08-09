package spot_web_api

import (
	common "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func New(apiKey, apiSecret, symbol string, sign signature.Sign, useTestNet ...bool) *common.WebApi {
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
	return common.New(apiKey, apiSecret, symbol, baseUrl, waHost, waPath, sign)
}
