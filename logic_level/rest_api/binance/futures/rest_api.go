package spot_rest_api

import (
	"sync"

	request "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/common/request"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func (ra *RestApi) Lock() {
	ra.mutex.Lock()
}

func (ra *RestApi) Unlock() {
	ra.mutex.Unlock()
}

func (ra *RestApi) NewOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/fapi/v1/Request", ra.sign)
}

func (ra *RestApi) QueryOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/fapi/v1/Request", ra.sign)
}

func (ra *RestApi) CancelOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/fapi/v1/Request", ra.sign)
}

func (ra *RestApi) CancelAllOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/fapi/v1/openOrders", ra.sign)
}

func (ra *RestApi) CancelReplaceOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/fapi/v1/Request/cancelReplace", ra.sign)
}

func (ra *RestApi) QueryOpenOrders() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/fapi/v1/openOrders", ra.sign)
}

func (ra *RestApi) QueryAllOrders() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/fapi/v1/allOrders", ra.sign)
}

func (ra *RestApi) ListenKey() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/fapi/v1/listenKey", ra.sign)
}

func (ra *RestApi) KeepAliveListenKey() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "PUT", ra.apiBaseUrl, "/api/v1/listenKey", ra.sign)
}

func (ra *RestApi) CloseListenKey() *request.Request {
	return request.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/api/v1/listenKey", ra.sign)
}

func New(apiKey, apiSecret string, symbol string, sign signature.Sign, useTestNet ...bool) (api *RestApi) {
	const (
		BaseAPIMainUrl    = "https://fapi.binance.com"
		BaseAPITestnetUrl = "https://testnet.binancefuture.com"
	)
	api = &RestApi{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		mutex:     &sync.Mutex{},
		sign:      sign,
	}
	if len(useTestNet) > 0 && useTestNet[0] {
		api.apiBaseUrl = BaseAPITestnetUrl
	} else {
		api.apiBaseUrl = BaseAPIMainUrl
	}
	return
}
