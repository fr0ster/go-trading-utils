package orders

import (
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

type (
	OrderType           string
	SideType            string
	TimeInForceType     string
	QuantityType        string
	OrderStatusType     string
	WorkingType         string
	PositionSideType    string
	CreateOrderResponse struct {
		Symbol           string          `json:"symbol"`        //
		OrderID          int64           `json:"orderId"`       //
		ClientOrderID    string          `json:"clientOrderId"` //
		Price            string          `json:"price"`         //
		OrigQuantity     string          `json:"origQty"`       //
		ExecutedQuantity string          `json:"executedQty"`   //
		Status           OrderStatusType `json:"status"`        //
		StopPrice        string          `json:"stopPrice"`     // please ignore when order type is TRAILING_STOP_MARKET
		TimeInForce      TimeInForceType `json:"timeInForce"`   //
		Type             OrderType       `json:"type"`          //
		Side             SideType        `json:"side"`          //
		UpdateTime       int64           `json:"updateTime"`    // update time
	}
	CancelOrderResponse struct {
		ClientOrderID    string           `json:"clientOrderId"`
		CumQuantity      string           `json:"cumQty"`
		CumQuote         string           `json:"cumQuote"`
		ExecutedQuantity string           `json:"executedQty"`
		OrderID          int64            `json:"orderId"`
		OrigQuantity     string           `json:"origQty"`
		Price            string           `json:"price"`
		ReduceOnly       bool             `json:"reduceOnly"`
		Side             SideType         `json:"side"`
		Status           OrderStatusType  `json:"status"`
		StopPrice        string           `json:"stopPrice"`
		Symbol           string           `json:"symbol"`
		TimeInForce      TimeInForceType  `json:"timeInForce"`
		Type             OrderType        `json:"type"`
		UpdateTime       int64            `json:"updateTime"`
		WorkingType      WorkingType      `json:"workingType"`
		ActivatePrice    string           `json:"activatePrice"`
		PriceRate        string           `json:"priceRate"`
		OrigType         string           `json:"origType"`
		PositionSide     PositionSideType `json:"positionSide"`
		PriceProtect     bool             `json:"priceProtect"`
	}
	Order struct {
		Symbol                  string           `json:"symbol"`
		OrderID                 int64            `json:"orderId"`
		ClientOrderID           string           `json:"clientOrderId"`
		Price                   string           `json:"price"`
		ReduceOnly              bool             `json:"reduceOnly"`
		OrigQuantity            string           `json:"origQty"`
		ExecutedQuantity        string           `json:"executedQty"`
		CumQuantity             string           `json:"cumQty"`
		CumQuote                string           `json:"cumQuote"`
		Status                  OrderStatusType  `json:"status"`
		TimeInForce             TimeInForceType  `json:"timeInForce"`
		Type                    OrderType        `json:"type"`
		Side                    SideType         `json:"side"`
		StopPrice               string           `json:"stopPrice"`
		Time                    int64            `json:"time"`
		UpdateTime              int64            `json:"updateTime"`
		WorkingType             WorkingType      `json:"workingType"`
		ActivatePrice           string           `json:"activatePrice"`
		PriceRate               string           `json:"priceRate"`
		AvgPrice                string           `json:"avgPrice"`
		OrigType                OrderType        `json:"origType"`
		PositionSide            PositionSideType `json:"positionSide"`
		PriceProtect            bool             `json:"priceProtect"`
		ClosePosition           bool             `json:"closePosition"`
		PriceMatch              string           `json:"priceMatch"`
		SelfTradePreventionMode string           `json:"selfTradePreventionMode"`
		GoodTillDate            int64            `json:"goodTillDate"`
	}
	CreateOrderFunction func(
		orderType OrderType,
		sideType SideType,
		timeInForce TimeInForceType,
		quantity items_types.QuantityType,
		closePosition bool,
		reduceOnly bool,
		price items_types.PriceType,
		stopPrice items_types.PriceType,
		activationPrice items_types.PriceType,
		callbackRate items_types.PricePercentType) (CreateOrderResponse, error)
	Orders struct {
		symbol              string
		stop                chan struct{}
		resetEvent          chan error
		timeOut             time.Duration
		startUserDataStream types.StreamFunction
		CreateOrder         CreateOrderFunction
		GetOpenOrders       func() ([]*Order, error)
		GetAllOrders        func() ([]*Order, error)
		GetOrder            func(orderID int64) (*Order, error)
		CancelOrder         func(orderID int64) (*CancelOrderResponse, error)
		CancelAllOrders     func() (err error)
	}
)

func (o *Orders) ResetEvent(err error) {
	o.resetEvent <- err
}

func (o *Orders) Symbol() string {
	return o.symbol
}

func New(
	symbol string,
	startUserDataStreamCreator func(*Orders) types.StreamFunction,
	createOrderCreator func(*Orders) CreateOrderFunction) {
	this := &Orders{
		symbol:     symbol,
		stop:       make(chan struct{}),
		resetEvent: make(chan error),
		timeOut:    1 * time.Hour,
	}
	if startUserDataStreamCreator != nil {
		this.startUserDataStream = startUserDataStreamCreator(this)
	}
	if createOrderCreator != nil {
		this.CreateOrder = createOrderCreator(this)
	}
}
