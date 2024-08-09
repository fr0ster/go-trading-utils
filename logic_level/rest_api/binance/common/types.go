package common_rest_api

import (
	"sync"

	rest_api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	RestApi struct {
		apiKey     string
		apiSecret  string
		symbol     string
		apiBaseUrl rest_api.ApiBaseUrl
		mutex      *sync.Mutex
		sign       signature.Sign
	}
)

func (wa *RestApi) Lock() {
	wa.mutex.Lock()
}

func (wa *RestApi) Unlock() {
	wa.mutex.Unlock()
}

// func (wa *RestApi) PlaceOrder() *order.Order {
// 	return newPlaceOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }

// func (wa *RestApi) CancelOrder() *order.Order {
// 	return newCancelOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }

// func (wa *RestApi) QueryOrder() *order.Order {
// 	return newQueryOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }

// func (wa *RestApi) CancelReplaceOrder() *order.Order {
// 	return newCancelReplaceOrder(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }

// func (wa *RestApi) QueryOpenOrders() *order.Order {
// 	return newQueryOpenOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }

// func (wa *RestApi) QueryAllOrders() *order.Order {
// 	return newQueryAllOrders(wa.apiKey, wa.symbol, wa.waHost, wa.waPath, wa.sign)
// }
