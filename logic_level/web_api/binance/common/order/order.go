package order

import (
	"encoding/json"
	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/fr0ster/turbo-restler/utils/signature"
	"github.com/fr0ster/turbo-restler/web_api"
)

type (
	Order struct {
		sign   signature.Sign
		waHost web_api.WsHost
		waPath web_api.WsPath
		method web_api.Method
		params *simplejson.Json
	}
)

func (o *Order) Set(name string, value interface{}) *Order {
	o.params.Set(name, value)
	return o
}

func (po *Order) Do() (order *simplejson.Json, err error) {
	response, err := web_api.CallWebAPI(web_api.WsHost(po.waHost), po.waPath, po.method, po.params, po.sign)
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

func New(apiKey, symbol string, method web_api.Method, waHost web_api.WsHost, waPath web_api.WsPath, sign signature.Sign) *Order {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &Order{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		params: simpleJson,
		method: method,
	}
}
