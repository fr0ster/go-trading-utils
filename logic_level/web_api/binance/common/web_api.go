package common_web_api

import (
	"fmt"
	"sync"

	"github.com/bitly/go-simplejson"
	order "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/order"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
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

// Функція для логіну
func (wa *WebApi) Logon() (result *LogonResult, err error) {
	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("apiKey", wa.apiKey)

	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.logon", params, nil)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*LogonResult)
	return
}

// Функція для логіну
func (wa *WebApi) Logout() (result *LogonResult, err error) {
	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.logout", nil, nil)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*LogonResult)
	return
}

// Функція для перевірки статусу сесії
func (wa *WebApi) Status() (result *LogonResult, err error) {
	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.status", nil, nil)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*LogonResult)
	return
}

func New(apiKey, apiSecret string, host web_api.WsHost, path web_api.WsPath, symbol string, sign signature.Sign) *WebApi {
	return &WebApi{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		symbol:    symbol,
		waHost:    host,
		waPath:    path,
		mutex:     &sync.Mutex{},
		sign:      sign,
	}
}
