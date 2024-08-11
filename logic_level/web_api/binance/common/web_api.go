package common_web_api

import (
	"fmt"
	"sync"

	"github.com/bitly/go-simplejson"
	request "github.com/fr0ster/go-trading-utils/logic_level/web_api/binance/common/request"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

func (wa *WebApi) Lock() {
	wa.mutex.Lock()
}

func (wa *WebApi) Unlock() {
	wa.mutex.Unlock()
}

func (wa *WebApi) PlaceRequest() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "Request.place", wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelRequest() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "Request.cancel", wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryRequest() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "Request.status", wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) CancelReplaceRequest() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "Request.cancelReplace", wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryOpenRequests() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "openRequests.status", wa.waHost, wa.waPath, wa.sign)
}

func (wa *WebApi) QueryAllRequests() *request.Request {
	return request.New(wa.apiKey, wa.symbol, "Request.allRequests", wa.waHost, wa.waPath, wa.sign)
}

// Функція для логіну
func (wa *WebApi) Logon() (result *LogonResult, err error) {
	// Створення параметрів запиту
	params := simplejson.New()
	params.Set("apiKey", wa.apiKey)

	response, err := web_api.CallWebAPI(wa.waHost, wa.waPath, "session.logon", params, wa.sign)
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
