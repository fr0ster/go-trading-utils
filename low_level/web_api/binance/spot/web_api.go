package spot_web_api

import (
	"sync"

	signature "github.com/fr0ster/go-trading-utils/low_level/common/signature"
)

func NewWebApi(apiKey, apiSecret, symbol, baseUrl string, sign signature.Sign, useTestNet ...bool) *WebApi {
	var (
		waHost string
		waPath string
	)
	if len(useTestNet) == 0 {
		useTestNet = append(useTestNet, false)
	}
	if useTestNet[0] {
		waHost = "testnet.binancefuture.com"
		waPath = "/ws-fapi/v1"
	} else {
		waHost = "ws-fapi.binance.com"
		waPath = "/ws-fapi/v1"
	}
	return &WebApi{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		symbol:     symbol,
		baseUrl:    baseUrl,
		useTestNet: useTestNet[0],
		waHost:     waHost,
		waPath:     waPath,
		mutex:      &sync.Mutex{},
		sign:       sign,
	}
}
