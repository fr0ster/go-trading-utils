package spot_web_api

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	PlaceOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}
	PlaceOrderResult struct {
		ClientOrderId string `json:"clientOrderId"`
		OrderId       int    `json:"orderId"`
		OrderListId   int    `json:"orderListId"`
		Symbol        string `json:"symbol"`
		TransactTime  int64  `json:"transactTime"`
	}
)

// / Функція для встановлення параметрів
func (po *PlaceOrder) Set(name string, value interface{}) *PlaceOrder {
	po.params.Set(name, value)
	return po
}

// Функція для розміщення ордера через WebSocket
func (po *PlaceOrder) Do() (order *PlaceOrderResult, err error) {
	response, err := web_api.CallWebAPI(po.waHost, po.waPath, po.method, po.params, po.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	order = response.Result.(*PlaceOrderResult)

	return
}

func newOrder(apiKey, symbol, waHost, waPath string, sign signature.Sign) *PlaceOrder {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &PlaceOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.place",
		params: simpleJson,
	}
}
