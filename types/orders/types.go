package orders

import (
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

type (
	CreateOrderResponse struct {
		Symbol           string                `json:"symbol"`        //
		OrderID          int64                 `json:"orderId"`       //
		ClientOrderID    string                `json:"clientOrderId"` //
		Price            string                `json:"price"`         //
		OrigQuantity     string                `json:"origQty"`       //
		ExecutedQuantity string                `json:"executedQty"`   //
		Status           types.OrderStatusType `json:"status"`        //
		StopPrice        string                `json:"stopPrice"`     // please ignore when order type is TRAILING_STOP_MARKET
		TimeInForce      types.TimeInForceType `json:"timeInForce"`   //
		Type             types.OrderType       `json:"type"`          //
		Side             types.SideType        `json:"side"`          //
		UpdateTime       int64                 `json:"updateTime"`    // update time
	}
	CancelOrderResponse struct {
		ClientOrderID    string                 `json:"clientOrderId"`
		CumQuantity      string                 `json:"cumQty"`
		CumQuote         string                 `json:"cumQuote"`
		ExecutedQuantity string                 `json:"executedQty"`
		OrderID          int64                  `json:"orderId"`
		OrigQuantity     string                 `json:"origQty"`
		Price            string                 `json:"price"`
		ReduceOnly       bool                   `json:"reduceOnly"`
		Side             types.SideType         `json:"side"`
		Status           types.OrderStatusType  `json:"status"`
		StopPrice        string                 `json:"stopPrice"`
		Symbol           string                 `json:"symbol"`
		TimeInForce      types.TimeInForceType  `json:"timeInForce"`
		Type             types.OrderType        `json:"type"`
		UpdateTime       int64                  `json:"updateTime"`
		WorkingType      types.WorkingType      `json:"workingType"`
		ActivatePrice    string                 `json:"activatePrice"`
		PriceRate        string                 `json:"priceRate"`
		OrigType         string                 `json:"origType"`
		PositionSide     types.PositionSideType `json:"positionSide"`
		PriceProtect     bool                   `json:"priceProtect"`
	}
	Order struct {
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
	CreateOrderFunction func(
		orderType types.OrderType,
		sideType types.SideType,
		timeInForce types.TimeInForceType,
		quantity items_types.QuantityType,
		closePosition bool,
		reduceOnly bool,
		price items_types.PriceType,
		stopPrice items_types.PriceType,
		activationPrice items_types.PriceType,
		callbackRate items_types.PricePercentType) (*CreateOrderResponse, error)
	OpenOrderFunction       func() ([]*Order, error)
	AllOrdersFunction       func() ([]*Order, error)
	GetOrderFunction        func(orderID int64) (*Order, error)
	CancelOrderFunction     func(orderID int64) (*CancelOrderResponse, error)
	CancelAllOrdersFunction func() (err error)
	Orders                  struct {
		symbol              string
		stop                chan struct{}
		isStartedStream     bool
		resetEvent          chan error
		timeOut             time.Duration
		startUserDataStream types.StreamFunction
		CreateOrder         CreateOrderFunction
		GetOpenOrders       OpenOrderFunction
		GetAllOrders        AllOrdersFunction
		GetOrder            GetOrderFunction
		CancelOrder         CancelOrderFunction
		CancelAllOrders     CancelAllOrdersFunction
	}
)
