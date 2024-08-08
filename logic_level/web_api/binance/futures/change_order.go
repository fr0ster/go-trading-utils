package futures_web_api

import (
	"fmt"

	common "github.com/fr0ster/turbo-restler/utils/json"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	CancelReplaceOrderParams struct {
		ApiKey                     string `json:"apiKey"`                     // YES
		CancelNewClientOrderId     string `json:"cancelNewClientOrderId"`     // NO: New ID for the canceled order. Automatically generated if not sent
		CancelOrderId              int    `json:"cancelOrderId"`              // YES: Cancel order by orderId
		CancelOrigClientOrderId    string `json:"cancelOrigClientOrderId"`    // YES: Cancel order by clientOrderId
		CancelReplaceMode          string `json:"cancelReplaceMode"`          // YES
		CancelRestrictions         string `json:"cancelRestrictions"`         // NO: Supported values: ONLY_NEW, ONLY_PARTIALLY_FILLED
		IcebergQty                 string `json:"icebergQty"`                 // NO
		NewClientOrderId           string `json:"newClientOrderId"`           // NO: Arbitrary unique ID among open orders. Automatically generated if not sent
		NewOrderRespType           string `json:"newOrderRespType"`           // NO: Select response format: ACK, RESULT, FULL
		OrderRateLimitExceededMode string `json:"orderRateLimitExceededMode"` // NO: Supported values: DO_NOTHING (default), CANCEL_ONLY
		Price                      string `json:"price"`                      // NO *
		Quantity                   string `json:"quantity"`                   // NO *
		QuoteOrderQty              string `json:"quoteOrderQty"`              // NO *
		RecvWindow                 int    `json:"recvWindow"`                 // NO: The value cannot be greater than 60000
		SelfTradePreventionMode    string `json:"selfTradePreventionMode"`    // NO: Supported values: EXPIRE_TAKER, EXPIRE_MAKER, EXPIRE_BOTH, NONE
		Side                       string `json:"side"`                       // YES: BUY or SELL
		Signature                  string `json:"signature"`                  // YES
		StopPrice                  string `json:"stopPrice"`                  // NO *
		StrategyId                 int    `json:"strategyId"`                 // NO: Arbitrary numeric value identifying the order within an order strategy
		StrategyType               int    `json:"strategyType"`               // NO: Arbitrary numeric value identifying the order strategy. Values smaller than 1000000 are reserved and cannot be used
		Symbol                     string `json:"symbol"`                     // YES
		TimeInForce                string `json:"timeInForce"`                // NO *
		Timestamp                  int64  `json:"timestamp"`                  // YES
		TrailingDelta              string `json:"trailingDelta"`              // NO *: See Trailing Stop order FAQ
		Type                       string `json:"type"`                       // YES
	}

	CancelReplaceOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *CancelReplaceOrderParams
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

// Функція для встановлення CancelOrderId
func (cro *CancelReplaceOrder) SetCancelOrderId(cancelOrderId int) *CancelReplaceOrder {
	cro.params.CancelOrderId = cancelOrderId
	return cro
}

// Функція для встановлення CancelOrigClientOrderId
func (cro *CancelReplaceOrder) SetCancelOrigClientOrderId(cancelOrigClientOrderId string) *CancelReplaceOrder {
	cro.params.CancelOrigClientOrderId = cancelOrigClientOrderId
	return cro
}

// Функція для встановлення CancelNewClientOrderId
func (cro *CancelReplaceOrder) SetCancelNewClientOrderId(cancelNewClientOrderId string) *CancelReplaceOrder {
	cro.params.CancelNewClientOrderId = cancelNewClientOrderId
	return cro
}

// Функція для встановлення TimeInForce
func (cro *CancelReplaceOrder) SetTimeInForce(timeInForce string) *CancelReplaceOrder {
	cro.params.TimeInForce = timeInForce
	return cro
}

// Функція для встановлення Price
func (cro *CancelReplaceOrder) SetPrice(price string) *CancelReplaceOrder {
	cro.params.Price = price
	return cro
}

// Функція для встановлення Quantity
func (cro *CancelReplaceOrder) SetQuantity(quantity string) *CancelReplaceOrder {
	cro.params.Quantity = quantity
	return cro
}

// Функція для встановлення QuoteOrderQty
func (cro *CancelReplaceOrder) SetQuoteOrderQty(quoteOrderQty string) *CancelReplaceOrder {
	cro.params.QuoteOrderQty = quoteOrderQty
	return cro
}

// Функція для встановлення NewClientOrderId
func (cro *CancelReplaceOrder) SetNewClientOrderId(newClientOrderId string) *CancelReplaceOrder {
	cro.params.NewClientOrderId = newClientOrderId
	return cro
}

// Функція для встановлення NewOrderRespType
func (cro *CancelReplaceOrder) SetNewOrderRespType(newOrderRespType string) *CancelReplaceOrder {
	cro.params.NewOrderRespType = newOrderRespType
	return cro
}

// Функція для встановлення StopPrice
func (cro *CancelReplaceOrder) SetStopPrice(stopPrice string) *CancelReplaceOrder {
	cro.params.StopPrice = stopPrice
	return cro
}

// Функція для встановлення TrailingDelta
func (cro *CancelReplaceOrder) SetTrailingDelta(trailingDelta string) *CancelReplaceOrder {
	cro.params.TrailingDelta = trailingDelta
	return cro
}

// Функція для встановлення IcebergQty
func (cro *CancelReplaceOrder) SetIcebergQty(icebergQty string) *CancelReplaceOrder {
	cro.params.IcebergQty = icebergQty
	return cro
}

// Функція для встановлення StrategyId
func (cro *CancelReplaceOrder) SetStrategyId(strategyId int) *CancelReplaceOrder {
	cro.params.StrategyId = strategyId
	return cro
}

// Функція для встановлення StrategyType
func (cro *CancelReplaceOrder) SetStrategyType(strategyType int) *CancelReplaceOrder {
	cro.params.StrategyType = strategyType
	return cro
}

// Функція для встановлення SelfTradePreventionMode
func (cro *CancelReplaceOrder) SetSelfTradePreventionMode(selfTradePreventionMode string) *CancelReplaceOrder {
	cro.params.SelfTradePreventionMode = selfTradePreventionMode
	return cro
}

// Функція для встановлення CancelReplaceMode
func (cro *CancelReplaceOrder) SetCancelReplaceMode(cancelReplaceMode string) *CancelReplaceOrder {
	cro.params.CancelReplaceMode = cancelReplaceMode
	return cro
}

// Функція для встановлення CancelRestrictions
func (cro *CancelReplaceOrder) SetCancelRestrictions(cancelRestrictions string) *CancelReplaceOrder {
	cro.params.CancelRestrictions = cancelRestrictions
	return cro
}

// Функція для встановлення OrderRateLimitExceededMode
func (cro *CancelReplaceOrder) SetOrderRateLimitExceededMode(orderRateLimitExceededMode string) *CancelReplaceOrder {
	cro.params.OrderRateLimitExceededMode = orderRateLimitExceededMode
	return cro
}

// Функція для встановлення RecvWindow
func (cro *CancelReplaceOrder) SetRecvWindow(recvWindow int) *CancelReplaceOrder {
	cro.params.RecvWindow = recvWindow
	return cro
}

// Функція для розміщення ордера через WebSocket
func (cro *CancelReplaceOrder) Do() (result *CancelReplaceOrderResult, err error) {
	// Створення параметрів запиту
	params, err := common.StructToUrlValues(cro.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response, err := web_api.CallWebAPI(cro.waHost, cro.waPath, cro.method, params, cro.sign)
	if err != nil {
		return
	}

	result = response.Result.(*CancelReplaceOrderResult)

	return
}

// Функція для створення нової структури CancelReplaceOrderParams
func newCancelReplaceOrder(apiKey, symbol, waHost, waPath string, sign signature.Sign) *CancelReplaceOrder {
	return &CancelReplaceOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.cancelReplace",
		params: &CancelReplaceOrderParams{
			ApiKey: apiKey,
			Symbol: symbol,
		},
	}
}
