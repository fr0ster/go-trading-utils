package spot_rest_api

import (
	"sync"

	"github.com/bitly/go-simplejson"
	order "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/common/order"
	rest_api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

func (ra *RestApi) Lock() {
	ra.mutex.Lock()
}

func (ra *RestApi) Unlock() {
	ra.mutex.Unlock()
}

func (ra *RestApi) NewOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/api/v3/order", ra.sign)
}

func (ra *RestApi) TestOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/order/test", ra.sign)
}

func (ra *RestApi) QueryOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/order", ra.sign)
}

func (ra *RestApi) CancelOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/api/v3/order", ra.sign)
}

func (ra *RestApi) CancelAllOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/api/v3/openOrders", ra.sign)
}

func (ra *RestApi) CancelReplaceOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/api/v3/order/cancelReplace", ra.sign)
}

func (ra *RestApi) QueryOpenOrders() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "GET /api/v3/openOrders", ra.sign)
}

func (ra *RestApi) QueryAllOrders() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/api/v3/allOrders", ra.sign)
}

func New(apiKey, apiSecret string, baseUrl rest_api.ApiBaseUrl, symbol string, sign signature.Sign, useTestNet ...bool) (api *RestApi) {
	const (
		BaseAPIMainUrl    = "https://api.binance.com"
		BaseAPITestnetUrl = "https://testnet.binance.vision"
	)
	api = &RestApi{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		symbol:     symbol,
		apiBaseUrl: baseUrl,
		mutex:      &sync.Mutex{},
		sign:       sign,
	}
	if len(useTestNet) > 0 && useTestNet[0] {
		api.apiBaseUrl = BaseAPITestnetUrl
	} else {
		api.apiBaseUrl = BaseAPIMainUrl
	}
	return
}
