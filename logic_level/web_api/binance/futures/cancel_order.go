package futures_web_api

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	CancelOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}

	CancelResult struct {
		Symbol                  string `json:"symbol"`
		OrigClientOrderId       string `json:"origClientOrderId"`
		OrderId                 int64  `json:"orderId"`
		OrderListId             int64  `json:"orderListId"`
		ClientOrderId           string `json:"clientOrderId"`
		TransactTime            int64  `json:"transactTime"`
		Price                   string `json:"price"`
		OrigQty                 string `json:"origQty"`
		ExecutedQty             string `json:"executedQty"`
		CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
		Status                  string `json:"status"`
		TimeInForce             string `json:"timeInForce"`
		Type                    string `json:"type"`
		Side                    string `json:"side"`
		StopPrice               string `json:"stopPrice,omitempty"`
		TrailingDelta           int    `json:"trailingDelta,omitempty"`
		IcebergQty              string `json:"icebergQty,omitempty"`
		StrategyId              int64  `json:"strategyId,omitempty"`
		StrategyType            int64  `json:"strategyType,omitempty"`
		SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty"`
	}
)

// Функція для встановлення
func (co *CancelOrder) Set(name string, value interface{}) *CancelOrder {
	co.params.Set(name, value)
	return co
}

// Функція для розміщення ордера через WebSocket
func (co *CancelOrder) Do() (result *CancelResult, err error) {
	response, err := web_api.CallWebAPI(co.waHost, co.waPath, co.method, co.params, co.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*CancelResult)
	return
}

// Функція для створення нової структури CancelOrderParams
func newCancelOrder(apiKey string, symbol, waHost, waPath string, sign signature.Sign) *CancelOrder {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &CancelOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.cancel",
		params: simpleJson,
	}
}
