package futures_rest_api

import common "github.com/fr0ster/turbo-restler/rest_api"

const (
	BaseAPIMainUrl    = "https://fapi.binance.com"
	BaseAPITestnetUrl = "https://testnet.binancefuture.com"
)

func GetAPIBaseUrl(useTestNet ...bool) (endpoint common.ApiBaseUrl) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseAPITestnetUrl
	} else {
		endpoint = BaseAPIMainUrl
	}
	return
}
