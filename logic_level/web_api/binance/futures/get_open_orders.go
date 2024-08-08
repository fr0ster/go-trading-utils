package futures_web_api

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	QueryOpenOrders struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}

	QueryOpenOrdersResults []QueryOpenOrdersResult

	QueryOpenOrdersResult struct {
		ClientOrderId           string `json:"clientOrderId"`
		CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
		ExecutedQty             string `json:"executedQty"`
		IcebergQty              string `json:"icebergQty"`
		IsWorking               bool   `json:"isWorking"`
		OrderId                 int64  `json:"orderId"`
		OrderListId             int    `json:"orderListId"`
		OrigQty                 string `json:"origQty"`
		OrigQuoteOrderQty       string `json:"origQuoteOrderQty"`
		Price                   string `json:"price"`
		SelfTradePreventionMode string `json:"selfTradePreventionMode"`
		Side                    string `json:"side"`
		Status                  string `json:"status"`
		StopPrice               string `json:"stopPrice"`
		Symbol                  string `json:"symbol"`
		Time                    int64  `json:"time"`
		TimeInForce             string `json:"timeInForce"`
		Type                    string `json:"type"`
		UpdateTime              int64  `json:"updateTime"`
		WorkingTime             int64  `json:"workingTime"`
	}
)

// Функція для встановлення параметрів
func (qoo *QueryOpenOrders) Set(name string, value interface{}) *QueryOpenOrders {
	qoo.params.Set(name, value)
	return qoo
}

// Функція для розміщення ордера через WebSocket
func (qoo *QueryOpenOrders) Do(side, orderType, timeInForce, price, quantity string) (result *QueryOpenOrdersResults, err error) {
	response, err := web_api.CallWebAPI(qoo.waHost, qoo.waPath, qoo.method, qoo.params, qoo.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*QueryOpenOrdersResults)
	return
}

func newQueryOpenOrders(apiKey, symbol, waHost, waPath string, sign signature.Sign) *QueryOpenOrders {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &QueryOpenOrders{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "openOrders.status",
		params: simpleJson,
	}
}
