package spot_rest_api

import (
	"sync"

	"github.com/bitly/go-simplejson"
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
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/api/v3/Request", ra.sign)
}

func (ra *RestApi) TestOrder() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/Request/test", ra.sign)
}

func (ra *RestApi) QueryOrder() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/Request", ra.sign)
}

func (ra *RestApi) CancelOrder() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/api/v3/Request", ra.sign)
}

func (ra *RestApi) CancelAllOrder() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/api/v3/openOrders", ra.sign)
}

func (ra *RestApi) CancelReplaceOrder() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/api/v3/Request/cancelReplace", ra.sign)
}

func (ra *RestApi) QueryOpenOrders() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/openOrders", ra.sign)
}

func (ra *RestApi) QueryAllOrders() *request.Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return request.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/allOrders", ra.sign)
}

func New(apiKey, apiSecret string, symbol string, sign signature.Sign, useTestNet ...bool) (api *RestApi) {
	const (
		BaseAPIMainUrl    = "https://api.binance.com"
		BaseAPITestnetUrl = "https://testnet.binance.vision"
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
