package spot_web_api

import (
	"fmt"

	"github.com/bitly/go-simplejson"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	QueryOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *simplejson.Json
	}

	QueryOrderResult struct {
		Symbol                  string `json:"symbol"`
		OrderId                 int64  `json:"orderId"`
		OrderListId             int64  `json:"orderListId"`
		ClientOrderId           string `json:"clientOrderId"`
		Price                   string `json:"price"`
		OrigQty                 string `json:"origQty"`
		ExecutedQty             string `json:"executedQty"`
		CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
		Status                  string `json:"status"`
		TimeInForce             string `json:"timeInForce"`
		Type                    string `json:"type"`
		Side                    string `json:"side"`
		StopPrice               string `json:"stopPrice"`
		TrailingDelta           int    `json:"trailingDelta,omitempty"`
		TrailingTime            int64  `json:"trailingTime,omitempty"`
		IcebergQty              string `json:"icebergQty"`
		Time                    int64  `json:"time"`
		UpdateTime              int64  `json:"updateTime"`
		IsWorking               bool   `json:"isWorking"`
		WorkingTime             int64  `json:"workingTime"`
		OrigQuoteOrderQty       string `json:"origQuoteOrderQty"`
		StrategyId              int64  `json:"strategyId,omitempty"`
		StrategyType            int64  `json:"strategyType,omitempty"`
		SelfTradePreventionMode string `json:"selfTradePreventionMode"`
		PreventedMatchId        int64  `json:"preventedMatchId,omitempty"`
		PreventedQuantity       string `json:"preventedQuantity,omitempty"`
	}
)

// Функція для встановлення параметрів
func (qo *QueryOrder) Set(name string, value interface{}) *QueryOrder {
	qo.params.Set(name, value)
	return qo
}

// Функція для розміщення ордера через WebSocket
func (qo *QueryOrder) Do(side, orderType, timeInForce, price, quantity string) (result *QueryOrderResult, err error) {
	response, err := web_api.CallWebAPI(qo.waHost, qo.waPath, qo.method, qo.params, qo.sign)
	if err != nil {
		return
	}

	if response.Status != 200 {
		err = fmt.Errorf("error request: %v", response.Error)
		return
	}

	result = response.Result.(*QueryOrderResult)

	return
}

func newQueryOrder(apiKey, symbol, waHost, waPath string, sign signature.Sign) *QueryOrder {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &QueryOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.status",
		params: simpleJson,
	}
}
