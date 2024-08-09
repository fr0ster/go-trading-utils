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
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}
)

func (o *Order) Set(name string, value interface{}) *Order {
	o.params.Set(name, value)
	return o
}

func (po *Order) Do() (order *simplejson.Json, err error) {
	response, err := web_api.CallWebAPI(po.waHost, po.waPath, po.method, po.params, po.sign)
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

func New(apiKey, symbol, method, waHost, waPath string, sign signature.Sign) *Order {
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
