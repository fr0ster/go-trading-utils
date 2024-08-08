package futures_web_api

import (
	"fmt"

	common "github.com/fr0ster/turbo-restler/utils/json"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	QueryOrderParams struct {
		ApiKey            string `json:"apiKey"`            // STRING, YES
		OrderId           int64  `json:"orderId"`           // LONG, YES
		OrigClientOrderId string `json:"origClientOrderId"` // STRING, NO
		Signature         string `json:"signature"`         // STRING, YES
		Symbol            string `json:"symbol"`            // STRING, YES
		RecvWindow        int    `json:"recvWindow"`        // INT, NO
		Timestamp         int64  `json:"timestamp"`         // LONG, YES
	}

	QueryOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *QueryOrderParams
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

// Функція для встановлення OrigClientOrderId
func (qo *QueryOrder) SetOrigClientOrderId(origClientOrderId string) *QueryOrder {
	qo.params.OrigClientOrderId = origClientOrderId
	return qo
}

// Функція для встановлення RecvWindow
func (qo *QueryOrder) SetRecvWindow(recvWindow int) *QueryOrder {
	qo.params.RecvWindow = recvWindow
	return qo
}

// Функція для встановлення OrderId
func (qo *QueryOrder) SetOrderId(orderId int64) *QueryOrder {
	qo.params.OrderId = orderId
	return qo
}

// Функція для розміщення ордера через WebSocket
func (qo *QueryOrder) Do(side, orderType, timeInForce, price, quantity string) (result *QueryOrderResult, err error) {
	// Створення параметрів запиту
	params, err := common.StructToParameterMap(qo.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response, err := web_api.CallWebAPI(qo.waHost, qo.waPath, qo.method, params, qo.sign)
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
	return &QueryOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.status",
		params: &QueryOrderParams{
			ApiKey: apiKey,
			Symbol: symbol,
		},
	}
}
