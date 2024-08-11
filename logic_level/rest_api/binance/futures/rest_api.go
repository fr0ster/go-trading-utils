package spot_rest_api

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bitly/go-simplejson"

	request "github.com/fr0ster/go-trading-utils/logic_level/rest_api/binance/common/request"
	signature "github.com/fr0ster/turbo-restler/utils/signature"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"
)

func (ra *RestApi) Lock() {
	ra.mutex.Lock()
}

func (ra *RestApi) Unlock() {
	ra.mutex.Unlock()
}

func (ra *RestApi) NewOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodPost, ra.apiBaseUrl, "/fapi/v1/Request", ra.parameters, ra.sign)
}

func (ra *RestApi) QueryOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodGet, ra.apiBaseUrl, "/fapi/v1/Request", ra.parameters, ra.sign)
}

func (ra *RestApi) CancelOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodDelete, ra.apiBaseUrl, "/fapi/v1/Request", ra.parameters, ra.sign)
}

func (ra *RestApi) CancelAllOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodDelete, ra.apiBaseUrl, "/fapi/v1/openOrders", ra.parameters, ra.sign)
}

func (ra *RestApi) CancelReplaceOrder() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodPost, ra.apiBaseUrl, "/fapi/v1/Request/cancelReplace", ra.parameters, ra.sign)
}

func (ra *RestApi) QueryOpenOrders() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodGet, ra.apiBaseUrl, "/fapi/v1/openOrders", ra.parameters, ra.sign)
}

func (ra *RestApi) QueryAllOrders() *request.Request {
	return request.New(ra.apiKey, ra.symbol, http.MethodGet, ra.apiBaseUrl, "/fapi/v1/allOrders", ra.parameters, ra.sign)
}

func (ra *RestApi) ListenKey() (listenKey string, err error) {
	response, err := rest_api.CallRestAPI(ra.apiBaseUrl, http.MethodPost, nil, "/api/v1/listenKey", ra.sign)
	if err != nil {
		err = fmt.Errorf("error calling API: %v", err)
		return
	}
	listenKey = response.Get("listenKey").MustString()
	return
}

func (ra *RestApi) KeepAliveListenKey() (err error) {
	_, err = rest_api.CallRestAPI(ra.apiBaseUrl, http.MethodPut, nil, "/api/v1/listenKey", ra.sign)
	return
}

func (ra *RestApi) CloseListenKey() (err error) {
	_, err = rest_api.CallRestAPI(ra.apiBaseUrl, http.MethodDelete, nil, "/api/v1/listenKey", ra.sign)
	return
}

func New(apiKey, apiSecret string, symbol string, sign signature.Sign, useTestNet ...bool) (api *RestApi) {
	const (
		BaseAPIMainUrl    = "https://fapi.binance.com"
		BaseAPITestnetUrl = "https://testnet.binancefuture.com"
	)
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	api = &RestApi{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		symbol:     symbol,
		mutex:      &sync.Mutex{},
		parameters: simpleJson,
		sign:       sign,
	}
	if len(useTestNet) > 0 && useTestNet[0] {
		api.apiBaseUrl = BaseAPITestnetUrl
	} else {
		api.apiBaseUrl = BaseAPIMainUrl
	}
	return
}
