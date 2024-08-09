package spot_rest_api

import api "github.com/fr0ster/turbo-restler/rest_api"

const (
	BaseAPIMainUrl    = "https://api.binance.com"
	BaseAPITestnetUrl = "https://testnet.binance.vision"
)

func GetAPIBaseUrl(useTestNet ...bool) (endpoint api.ApiBaseUrl) {
	if len(useTestNet) > 0 && useTestNet[0] {
		endpoint = BaseAPITestnetUrl
	} else {
		endpoint = BaseAPIMainUrl
	}
	return
}
