package spot_web_api

import (
	"sync"

	order "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/order"
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

func (wa *WebApi) PlaceOrder() *order.Order {
	return newPlaceOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelOrder() *order.Order {
	return newCancelOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryOrder() *order.Order {
	return newQueryOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelReplaceOrder() *order.Order {
	return newCancelReplaceOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryOpenOrders() *order.Order {
	return newQueryOpenOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryAllOrders() *order.Order {
	return newQueryAllOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
}

func newWebApi(apiKey, apiSecret, symbol, baseUrl, waHost, waPath string, sign signature.Sign) *WebApi {
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
