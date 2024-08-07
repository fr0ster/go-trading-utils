package spot_web_api

import (
	"sync"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	WebApi struct {
		apiKey    string
		apiSecret string
		symbol    string
		baseUrl   string
		waHost    string
		waPath    string
		mutex     *sync.Mutex
		sign      signature.Sign
	}
)

func (wa *WebApi) Lock() {
	wa.mutex.Lock()
}

func (wa *WebApi) Unlock() {
	wa.mutex.Unlock()
}

func (wa *WebApi) PlaceOrder() (response *PlaceOrder) {
	return newOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelOrder() (response *CancelOrder) {
	return newCancelOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryOrder() (response *QueryOrder) {
	return newQueryOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelReplaceOrder() (response *CancelReplaceOrder) {
	return newCancelReplaceOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryOpenOrders() (response *QueryOpenOrders) {
	return newQueryOpenOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryAllOrders() (response *QueryAllOrders) {
	return newQueryAllOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func newSpotWebApi(apiKey, apiSecret, symbol, baseUrl, waHost, waPath string, sign signature.Sign) *WebApi {
	return &WebApi{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		baseUrl:   baseUrl,
		waHost:    waHost,
		waPath:    waPath,
		mutex:     &sync.Mutex{},
		sign:      sign,
	}
}
