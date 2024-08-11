package order

import (
	"encoding/json"
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/fr0ster/turbo-restler/utils/signature"
	"github.com/fr0ster/turbo-restler/web_api"
)

type (
	Request struct {
		sign   signature.Sign
		waHost web_api.WsHost
		waPath web_api.WsPath
		method web_api.Method
		params *simplejson.Json
	}
)

func (rq *Request) Set(name string, value interface{}) *Request {
	rq.params.Set(name, value)
	return rq
}

func (rq *Request) Do() (order *simplejson.Json, err error) {
	response, err := web_api.CallWebAPI(web_api.WsHost(rq.waHost), rq.waPath, rq.method, rq.params, rq.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	bytes, err := json.Marshal(response.Result)
	if err != nil {
		return
	}
	order, err = simplejson.NewJson(bytes)
	return
}

func New(apiKey, symbol string, method web_api.Method, waHost web_api.WsHost, waPath web_api.WsPath, sign signature.Sign) *Request {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &Request{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		params: simpleJson,
		method: method,
	}
}
