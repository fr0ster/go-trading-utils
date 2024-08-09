package spot_rest_api

import (
	"sync"

	"github.com/bitly/go-simplejson"
	order "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/common/order"
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
	return order.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/fapi/v1/order", ra.sign)
}

func (ra *RestApi) QueryOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/fapi/v1/order", ra.sign)
}

func (ra *RestApi) CancelOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/fapi/v1/order", ra.sign)
}

func (ra *RestApi) CancelAllOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "DELETE", ra.apiBaseUrl, "/fapi/v1/openOrders", ra.sign)
}

func (ra *RestApi) CancelReplaceOrder() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "POST", ra.apiBaseUrl, "/fapi/v1/order/cancelReplace", ra.sign)
}

func (ra *RestApi) QueryOpenOrders() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "GET /fapi/v1/openOrders", ra.sign)
}

func (ra *RestApi) QueryAllOrders() *order.Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", ra.apiKey)
	simpleJson.Set("symbol", ra.symbol)
	return order.New(ra.apiKey, ra.symbol, "GET", ra.apiBaseUrl, "/fapi/v1/allOrders", ra.sign)
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
