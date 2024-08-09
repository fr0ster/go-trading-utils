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
	// QueryOrderResponse struct {
	// 	OrderID int64  `json:"orderId"`
	// 	Symbol  string `json:"symbol"`
	// 	Status  string `json:"status"`
	// }
	// OrderStatusType define order status type
	QueryOrderResponse struct {
		Symbol                   string                `json:"symbol"`
		OrderID                  int64                 `json:"orderId"`
		OrderListId              int64                 `json:"orderListId"`
		ClientOrderID            string                `json:"clientOrderId"`
		Price                    string                `json:"price"`
		OrigQuantity             string                `json:"origQty"`
		ExecutedQuantity         string                `json:"executedQty"`
		CummulativeQuoteQuantity string                `json:"cummulativeQuoteQty"`
		Status                   types.OrderStatusType `json:"status"`
		TimeInForce              types.TimeInForceType `json:"timeInForce"`
		Type                     types.OrderType       `json:"type"`
		Side                     types.SideType        `json:"side"`
		StopPrice                string                `json:"stopPrice"`
		IcebergQuantity          string                `json:"icebergQty"`
		Time                     int64                 `json:"time"`
		UpdateTime               int64                 `json:"updateTime"`
		IsWorking                bool                  `json:"isWorking"`
		IsIsolated               bool                  `json:"isIsolated"`
		OrigQuoteOrderQuantity   string                `json:"origQuoteOrderQty"`
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

	CreateOrderResponse struct {
		Symbol                   string `json:"symbol"`
		OrderID                  int64  `json:"orderId"`
		ClientOrderID            string `json:"clientOrderId"`
		TransactTime             int64  `json:"transactTime"`
		Price                    string `json:"price"`
		OrigQuantity             string `json:"origQty"`
		ExecutedQuantity         string `json:"executedQty"`
		CummulativeQuoteQuantity string `json:"cummulativeQuoteQty"`
		IsIsolated               bool   `json:"isIsolated"` // for isolated margin

		Status      types.OrderStatusType `json:"status"`
		TimeInForce types.TimeInForceType `json:"timeInForce"`
		Type        types.OrderType       `json:"type"`
		Side        types.SideType        `json:"side"`

		// for order response is set to FULL
		Fills                 []*Fill `json:"fills"`
		MarginBuyBorrowAmount string  `json:"marginBuyBorrowAmount"` // for margin
		MarginBuyBorrowAsset  string  `json:"marginBuyBorrowAsset"`
	}

	// Fill may be returned in an array of fills in a CreateOrderResponse.
	Fill struct {
		TradeID         int64  `json:"tradeId"`
		Price           string `json:"price"`
		Quantity        string `json:"qty"`
		Commission      string `json:"commission"`
		CommissionAsset string `json:"commissionAsset"`
	}
)
