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
	PlaceOrderParams struct {
		ApiKey                  string  `json:"apiKey"`                  // STRING, YES
		IcebergQty              float64 `json:"icebergQty"`              // DECIMAL, NO
		NewClientOrderId        string  `json:"newClientOrderId"`        // STRING, NO
		NewOrderRespType        string  `json:"newOrderRespType"`        // ENUM, NO
		Price                   float64 `json:"price"`                   // DECIMAL, NO
		Quantity                float64 `json:"quantity"`                // DECIMAL, NO
		QuoteOrderQty           float64 `json:"quoteOrderQty"`           // DECIMAL, NO
		RecvWindow              int     `json:"recvWindow"`              // INT, NO
		SelfTradePreventionMode string  `json:"selfTradePreventionMode"` // ENUM, NO
		Side                    string  `json:"side"`                    // ENUM, YES
		Signature               string  `json:"signature"`               // STRING, YES
		StopPrice               float64 `json:"stopPrice"`               // DECIMAL, NO
		StrategyId              int     `json:"strategyId"`              // INT, NO
		StrategyType            int     `json:"strategyType"`            // INT, NO
		Symbol                  string  `json:"symbol"`                  // STRING, YES
		TimeInForce             string  `json:"timeInForce"`             // ENUM, NO
		Timestamp               int64   `json:"timestamp"`               // INT, YES
		TrailingDelta           int     `json:"trailingDelta"`           // INT, NO
		Type                    string  `json:"type"`                    // ENUM, YES
	}
	PlaceOrderResult struct {
		ClientOrderId string `json:"clientOrderId"`
		OrderId       int    `json:"orderId"`
		OrderListId   int    `json:"orderListId"`
		Symbol        string `json:"symbol"`
		TransactTime  int64  `json:"transactTime"`
	}
	PlaceOrder struct {
		sign   signature.Sign
		waHost string
		waPath string
		method string
		params *PlaceOrderParams
	}
)

func (po *PlaceOrder) SetSide(side string) *PlaceOrder {
	po.params.Side = side
	return po
}

func (po *PlaceOrder) SetIcebergQty(icebergQty float64) *PlaceOrder {
	po.params.IcebergQty = icebergQty
	return po
}

func (po *PlaceOrder) SetNewClientOrderId(newClientOrderId string) *PlaceOrder {
	po.params.NewClientOrderId = newClientOrderId
	return po
}

func (po *PlaceOrder) SetNewOrderRespType(newOrderRespType string) *PlaceOrder {
	po.params.NewOrderRespType = newOrderRespType
	return po
}

func (po *PlaceOrder) SetPrice(price float64) *PlaceOrder {
	po.params.Price = price
	return po
}

func (po *PlaceOrder) SetQuantity(quantity float64) *PlaceOrder {
	po.params.Quantity = quantity
	return po
}

func (po *PlaceOrder) SetQuoteOrderQty(quoteOrderQty float64) *PlaceOrder {
	po.params.QuoteOrderQty = quoteOrderQty
	return po
}

func (po *PlaceOrder) SetRecvWindow(recvWindow int) *PlaceOrder {
	po.params.RecvWindow = recvWindow
	return po
}

func (po *PlaceOrder) SetSelfTradePreventionMode(selfTradePreventionMode string) *PlaceOrder {
	po.params.SelfTradePreventionMode = selfTradePreventionMode
	return po
}

func (po *PlaceOrder) SetSignature(signature string) *PlaceOrder {
	po.params.Signature = signature
	return po
}

func (po *PlaceOrder) SetStopPrice(stopPrice float64) *PlaceOrder {
	po.params.StopPrice = stopPrice
	return po
}

func (po *PlaceOrder) SetStrategyId(strategyId int) *PlaceOrder {
	po.params.StrategyId = strategyId
	return po
}

func (po *PlaceOrder) SetStrategyType(strategyType int) *PlaceOrder {
	po.params.StrategyType = strategyType
	return po
}

func (po *PlaceOrder) SetTimeInForce(timeInForce string) *PlaceOrder {
	po.params.TimeInForce = timeInForce
	return po
}

func (po *PlaceOrder) SetTimestamp(timestamp int64) *PlaceOrder {
	po.params.Timestamp = timestamp
	return po
}

func (po *PlaceOrder) SetTrailingDelta(trailingDelta int) *PlaceOrder {
	po.params.TrailingDelta = trailingDelta
	return po
}

// Функція для розміщення ордера через WebSocket
func (po *PlaceOrder) Do() (order *PlaceOrderResult, err error) {
	// Перетворення структури в строку
	po.params.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	message, err := common.StructToQueryString(po.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	po.params.Signature = po.sign.CreateSignature(message)

	response, err := web_api.CallWebAPI(po.waHost, po.waPath, po.method, po.params)
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
	return &PlaceOrder{
		sign:   sign,
		waHost: waHost,
		waPath: waPath,
		method: "order.place",
		params: &PlaceOrderParams{
			ApiKey: apiKey,
			Symbol: symbol,
		},
	}
}
