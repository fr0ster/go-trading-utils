package futures_web_api

import (
	"fmt"

	common "github.com/fr0ster/turbo-restler/utils/json"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
)

// Структура для параметрів запиту
type (
	PlaceOrderParams struct {
		ActivationPrice         float64 `json:"activationPrice"`         // NO: Used with TRAILING_STOP_MARKET orders, default as the latest price(supporting different workingType)
		ApiKey                  string  `json:"apiKey"`                  // YES
		CallbackRate            float64 `json:"callbackRate"`            // NO: Used with TRAILING_STOP_MARKET orders, min 0.1, max 5 where 1 for 1%
		ClosePosition           string  `json:"closePosition"`           // NO: true, false；Close-All，used with STOP_MARKET or TAKE_PROFIT_MARKET.
		GoodTillDate            int64   `json:"goodTillDate"`            // NO: order cancel time for timeInForce GTD, mandatory when timeInforce set to GTD; order the timestamp only retains second-level precision, ms part will be ignored; The goodTillDate timestamp must be greater than the current time plus 600 seconds and smaller than 253402300799000
		NewClientOrderId        string  `json:"newClientOrderId"`        // NO: A unique id among open orders. Automatically generated if not sent. Can only be string following the rule: ^[\.A-Z\:/a-z0-9_-]{1,36}$
		NewOrderRespType        string  `json:"newOrderRespType"`        // NO: "ACK", "RESULT", default "ACK"
		PositionSide            string  `json:"positionSide"`            // NO: Default BOTH for One-way Mode ; LONG or SHORT for Hedge Mode. It must be sent in Hedge Mode.
		Price                   float64 `json:"price"`                   // NO
		PriceMatch              string  `json:"priceMatch"`              // NO: only available for LIMIT/STOP/TAKE_PROFIT order; can be set to OPPONENT/ OPPONENT_5/ OPPONENT_10/ OPPONENT_20: /QUEUE/ QUEUE_5/ QUEUE_10/ QUEUE_20; Can't be passed together with price
		PriceProtect            string  `json:"priceProtect"`            // NO: "TRUE" or "FALSE", default "FALSE". Used with STOP/STOP_MARKET or TAKE_PROFIT/TAKE_PROFIT_MARKET orders.
		Quantity                float64 `json:"quantity"`                // NO: Cannot be sent with closePosition=true(Close-All)
		RecvWindow              int64   `json:"recvWindow"`              // NO
		ReduceOnly              string  `json:"reduceOnly"`              // NO: "true" or "false". default "false". Cannot be sent in Hedge Mode; cannot be sent with closePosition=true
		SelfTradePreventionMode string  `json:"selfTradePreventionMode"` // NO: NONE:No STP / EXPIRE_TAKER:expire taker order when STP triggers/ EXPIRE_MAKER:expire taker order when STP triggers/ EXPIRE_BOTH:expire both orders when STP triggers; default NONE
		Side                    string  `json:"side"`                    // YES
		Signature               string  `json:"signature"`               // YES
		StopPrice               float64 `json:"stopPrice"`               // NO: Used with STOP/STOP_MARKET or TAKE_PROFIT/TAKE_PROFIT_MARKET orders.
		Symbol                  string  `json:"symbol"`                  // YES
		Timestamp               int64   `json:"timestamp"`               // YES
		TimeInForce             string  `json:"timeInForce"`             // NO
		Type                    string  `json:"type"`                    // YES
		WorkingType             string  `json:"workingType"`             // NO: stopPrice triggered by: "MARK_PRICE", "CONTRACT_PRICE". Default "CONTRACT_PRICE"
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

func (po *PlaceOrder) SetActivationPrice(activationPrice float64) *PlaceOrder {
	po.params.ActivationPrice = activationPrice
	return po
}

func (po *PlaceOrder) SetApiKey(apiKey string) *PlaceOrder {
	po.params.ApiKey = apiKey
	return po
}

func (po *PlaceOrder) SetCallbackRate(callbackRate float64) *PlaceOrder {
	po.params.CallbackRate = callbackRate
	return po
}

func (po *PlaceOrder) SetClosePosition(closePosition string) *PlaceOrder {
	po.params.ClosePosition = closePosition
	return po
}

func (po *PlaceOrder) SetGoodTillDate(goodTillDate int64) *PlaceOrder {
	po.params.GoodTillDate = goodTillDate
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

func (po *PlaceOrder) SetPositionSide(positionSide string) *PlaceOrder {
	po.params.PositionSide = positionSide
	return po
}

func (po *PlaceOrder) SetPrice(price float64) *PlaceOrder {
	po.params.Price = price
	return po
}

func (po *PlaceOrder) SetPriceMatch(priceMatch string) *PlaceOrder {
	po.params.PriceMatch = priceMatch
	return po
}

func (po *PlaceOrder) SetPriceProtect(priceProtect string) *PlaceOrder {
	po.params.PriceProtect = priceProtect
	return po
}

func (po *PlaceOrder) SetQuantity(quantity float64) *PlaceOrder {
	po.params.Quantity = quantity
	return po
}

func (po *PlaceOrder) SetRecvWindow(recvWindow int64) *PlaceOrder {
	po.params.RecvWindow = recvWindow
	return po
}

func (po *PlaceOrder) SetReduceOnly(reduceOnly string) *PlaceOrder {
	po.params.ReduceOnly = reduceOnly
	return po
}

func (po *PlaceOrder) SetSelfTradePreventionMode(selfTradePreventionMode string) *PlaceOrder {
	po.params.SelfTradePreventionMode = selfTradePreventionMode
	return po
}

func (po *PlaceOrder) SetSide(side string) *PlaceOrder {
	po.params.Side = side
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

func (po *PlaceOrder) SetSymbol(symbol string) *PlaceOrder {
	po.params.Symbol = symbol
	return po
}

func (po *PlaceOrder) SetTimestamp(timestamp int64) *PlaceOrder {
	po.params.Timestamp = timestamp
	return po
}

func (po *PlaceOrder) SetTimeInForce(timeInForce string) *PlaceOrder {
	po.params.TimeInForce = timeInForce
	return po
}

func (po *PlaceOrder) SetType(orderType string) *PlaceOrder {
	po.params.Type = orderType
	return po
}

func (po *PlaceOrder) SetWorkingType(workingType string) *PlaceOrder {
	po.params.WorkingType = workingType
	return po
}

// Функція для розміщення ордера через WebSocket
func (po *PlaceOrder) Do() (order *PlaceOrderResult, err error) {
	// Перетворення структури в строку
	params, err := common.StructToParameterMap(po.params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response, err := web_api.CallWebAPI(po.waHost, po.waPath, po.method, params, po.sign)
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
