package futures_web_api

import (
	"github.com/bitly/go-simplejson"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	CancelReplaceOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}

	CancelReplaceOrderResult struct {
		CancelResult     string                 `json:"cancelResult"`
		NewOrderResult   string                 `json:"newOrderResult"`
		CancelResponse   CancelNewOrderResponse `json:"cancelResponse"`
		NewOrderResponse CancelNewOrderResponse `json:"newOrderResponse"`
	}

	CancelNewOrderResponse struct {
		Symbol                  string `json:"symbol"`
		OrigClientOrderId       string `json:"origClientOrderId"`
		OrderId                 int64  `json:"orderId"`
		OrderListId             int    `json:"orderListId"`
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
		SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	}
)

// Функція для встановлення
func (cro *CancelReplaceOrder) Set(name string, value interface{}) *CancelReplaceOrder {
	cro.params.Set(name, value)
	return cro
}

// Функція для розміщення ордера через WebSocket
func (cro *CancelReplaceOrder) Do() (result *CancelReplaceOrderResult, err error) {
	response, err := web_api.CallWebAPI(cro.waHost, cro.waPath, cro.method, cro.params, cro.sign)
	if err != nil {
		return
	}

	result = response.Result.(*CancelReplaceOrderResult)

	return
}

// Функція для створення нової структури CancelReplaceOrderParams
func newCancelReplaceOrder(apiKey, symbol, waHost, waPath string, sign signature.Sign) *CancelReplaceOrder {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &CancelReplaceOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.cancelReplace",
		params: simpleJson,
	}
}
