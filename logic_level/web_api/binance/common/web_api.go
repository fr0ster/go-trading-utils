package spot_web_api

import (
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func New(apiKey, apiSecret, baseUrl, host, path, symbol string, sign signature.Sign, useTestNet ...bool) *WebApi {
	return newWebApi(apiKey, apiSecret, symbol, baseUrl, host, path, sign)
}
