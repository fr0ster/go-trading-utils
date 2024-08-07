package futures_web_api

import (
	"fmt"
	"time"

	common "github.com/fr0ster/turbo-restler/utils/json"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	CancelOrderParams struct {
		ApiKey             string `json:"apiKey"`
		CancelRestrictions string `json:"cancelRestrictions,omitempty"`
		NewClientOrderId   string `json:"newClientOrderId,omitempty"`
		OrderId            int64  `json:"orderId"`
		OrigClientOrderId  string `json:"origClientOrderId,omitempty"`
		RecvWindow         int    `json:"recvWindow,omitempty"`
		Signature          string `json:"signature"`
		Symbol             string `json:"symbol"`
		Timestamp          int64  `json:"timestamp"`
	}

	CancelOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *CancelOrderParams
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

// Функція для встановлення CancelRestrictions
func (co *CancelOrder) SetCancelRestrictions(cancelRestrictions string) *CancelOrder {
	co.params.CancelRestrictions = cancelRestrictions
	return co
}

// Функція для встановлення NewClientOrderId
func (co *CancelOrder) SetNewClientOrderId(newClientOrderId string) *CancelOrder {
	co.params.NewClientOrderId = newClientOrderId
	return co
}

// Функція для встановлення OrigClientOrderId
func (co *CancelOrder) SetOrigClientOrderId(origClientOrderId string) *CancelOrder {
	co.params.OrigClientOrderId = origClientOrderId
	return co
}

// Функція для встановлення RecvWindow
func (co *CancelOrder) SetRecvWindow(recvWindow int) *CancelOrder {
	co.params.RecvWindow = recvWindow
	return co
}

// Функція для розміщення ордера через WebSocket
func (co *CancelOrder) CancelOrder(orderId int64, timeInForce string) (result *CancelResult, err error) {
	// Створення параметрів запиту
	co.params.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	// Перетворення структури в строку
	message, err := common.StructToQueryString(co.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	co.params.Signature = co.sign.CreateSignature(message)

	response, err := web_api.CallWebAPI(co.waHost, co.waPath, co.method, co.params)
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
	return &CancelOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.cancel",
		params: &CancelOrderParams{
			ApiKey: apiKey,
			Symbol: symbol,
		},
	}
}
