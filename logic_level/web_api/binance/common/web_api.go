package spot_web_api

import (
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

func New(apiKey, apiSecret string, host web_api.WsHost, path web_api.WsPath, symbol string, sign signature.Sign, useTestNet ...bool) *WebApi {
	return newWebApi(apiKey, apiSecret, symbol, host, path, sign)
}
