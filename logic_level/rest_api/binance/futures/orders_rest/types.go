package orders_rest

import (
	"sync"

	"github.com/fr0ster/go-trading-utils/types"
	common "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
)

type (
	// Структура для параметрів запиту
	SpotOrderRequest struct {
		Symbol      string  `json:"symbol"`
		Side        string  `json:"side"`
		Type        string  `json:"type"`
		Quantity    float64 `json:"quantity"`
		Price       float64 `json:"price,omitempty"`
		TimeInForce string  `json:"timeInForce,omitempty"`
		Timestamp   int64   `json:"timestamp"`
		Signature   string  `json:"signature"`
	}

	// Структура для відповіді API
	QueryOrderResponse struct {
		Symbol                  string                 `json:"symbol"`
		OrderID                 int64                  `json:"orderId"`
		ClientOrderID           string                 `json:"clientOrderId"`
		Price                   string                 `json:"price"`
		ReduceOnly              bool                   `json:"reduceOnly"`
		OrigQuantity            string                 `json:"origQty"`
		ExecutedQuantity        string                 `json:"executedQty"`
		CumQuantity             string                 `json:"cumQty"`
		CumQuote                string                 `json:"cumQuote"`
		Status                  types.OrderStatusType  `json:"status"`
		TimeInForce             types.TimeInForceType  `json:"timeInForce"`
		Type                    types.OrderType        `json:"type"`
		Side                    types.SideType         `json:"side"`
		StopPrice               string                 `json:"stopPrice"`
		Time                    int64                  `json:"time"`
		UpdateTime              int64                  `json:"updateTime"`
		WorkingType             types.WorkingType      `json:"workingType"`
		ActivatePrice           string                 `json:"activatePrice"`
		PriceRate               string                 `json:"priceRate"`
		AvgPrice                string                 `json:"avgPrice"`
		OrigType                types.OrderType        `json:"origType"`
		PositionSide            types.PositionSideType `json:"positionSide"`
		PriceProtect            bool                   `json:"priceProtect"`
		ClosePosition           bool                   `json:"closePosition"`
		PriceMatch              string                 `json:"priceMatch"`
		SelfTradePreventionMode string                 `json:"selfTradePreventionMode"`
		GoodTillDate            int64                  `json:"goodTillDate"`
	}
	// Структура для обробника ордерів
	Orders struct {
		apiKey    string
		apiSecret string
		symbol    string
		baseUrl   common.ApiBaseUrl
		mutex     *sync.Mutex
		sign      signature.Sign
	}

	// CreateOrderResponse define create order response
	CreateOrderResponse struct {
		Symbol                  string                 `json:"symbol"`                      //
		OrderID                 int64                  `json:"orderId"`                     //
		ClientOrderID           string                 `json:"clientOrderId"`               //
		Price                   string                 `json:"price"`                       //
		OrigQuantity            string                 `json:"origQty"`                     //
		ExecutedQuantity        string                 `json:"executedQty"`                 //
		CumQuote                string                 `json:"cumQuote"`                    //
		ReduceOnly              bool                   `json:"reduceOnly"`                  //
		Status                  types.OrderStatusType  `json:"status"`                      //
		StopPrice               string                 `json:"stopPrice"`                   // please ignore when order type is TRAILING_STOP_MARKET
		TimeInForce             types.TimeInForceType  `json:"timeInForce"`                 //
		Type                    types.OrderType        `json:"type"`                        //
		Side                    types.SideType         `json:"side"`                        //
		UpdateTime              int64                  `json:"updateTime"`                  // update time
		WorkingType             types.WorkingType      `json:"workingType"`                 //
		ActivatePrice           string                 `json:"activatePrice"`               // activation price, only return with TRAILING_STOP_MARKET order
		PriceRate               string                 `json:"priceRate"`                   // callback rate, only return with TRAILING_STOP_MARKET order
		AvgPrice                string                 `json:"avgPrice"`                    //
		PositionSide            types.PositionSideType `json:"positionSide"`                //
		ClosePosition           bool                   `json:"closePosition"`               // if Close-All
		PriceProtect            bool                   `json:"priceProtect"`                // if conditional order trigger is protected
		PriceMatch              string                 `json:"priceMatch"`                  // price match mode
		SelfTradePreventionMode string                 `json:"selfTradePreventionMode"`     // self trading preventation mode
		GoodTillDate            int64                  `json:"goodTillDate"`                // order pre-set auto cancel time for TIF GTD order
		CumQty                  string                 `json:"cumQty"`                      //
		OrigType                types.OrderType        `json:"origType"`                    //
		RateLimitOrder10s       string                 `json:"rateLimitOrder10s,omitempty"` //
		RateLimitOrder1m        string                 `json:"rateLimitOrder1m,omitempty"`  //
	}
)
