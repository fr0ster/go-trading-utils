package common_rest_api

import (
	"sync"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func New(apiKey, apiSecret string, baseUrl rest_api.ApiBaseUrl, symbol string, sign signature.Sign) *RestApi {
	return &RestApi{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		symbol:     symbol,
		apiBaseUrl: baseUrl,
		mutex:      &sync.Mutex{},
		sign:       sign,
	}
}
