package types

import (
	"time"

	"github.com/adshao/go-binance/v2"
)

type (
	OrderSide        string
	OrderType        string
	OrderStatusType  string
	DepthSide        string
	SideType         string
	PositionSideType string
	TimeInForceType  string
	WorkingType      string
	NewOrderRespType string
	DepthLevels      struct {
		Price    float64
		Side     DepthSide
		Quantity float64
	}
	SpotCreateOrderService struct {
		Side            SideType
		OrderType       OrderType
		Quantity        float64
		QuoteOrderQty   float64
		Price           float64
		StopPrice       float64
		TrailingDelta   float64
		IcebergQuantity float64
	}
	FuturesCreateOrderService struct {
		Side             SideType
		PositionSide     *PositionSideType
		OrderType        OrderType
		TimeInForce      *TimeInForceType
		Quantity         string
		ReduceOnly       *bool
		Price            *string
		NewClientOrderID *string
		StopPrice        *string
		WorkingType      *WorkingType
		ActivationPrice  *string
		CallbackRate     *string
		PriceProtect     *bool
		NewOrderRespType NewOrderRespType
		ClosePosition    *bool
	}
	Config struct {
		AccountType       binance.AccountType
		Symbol            string
		Balance           float64
		CalculatedBalance float64
		Quantity          float64
		Value             float64
		BoundQuantity     float64
	}
	Log struct {
		Timestamp time.Time
		Message   string
	}
)

const (
	DepthSideAsk                   DepthSide       = "ASK"
	DepthSideBid                   DepthSide       = "BID"
	SideTypeBuy                    OrderSide       = "BUY"
	SideTypeSell                   OrderSide       = "SELL"
	OrderTypeLimit                 OrderType       = "LIMIT"
	OrderTypeMarket                OrderType       = "MARKET"
	OrderTypeLimitMaker            OrderType       = "LIMIT_MAKER"
	OrderTypeStopLoss              OrderType       = "STOP_LOSS"
	OrderTypeStopLossLimit         OrderType       = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit            OrderType       = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit       OrderType       = "TAKE_PROFIT_LIMIT"
	OrderStatusTypeNew             OrderStatusType = "NEW"
	OrderStatusTypePartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusTypeFilled          OrderStatusType = "FILLED"
	OrderStatusTypeCanceled        OrderStatusType = "CANCELED"
	OrderStatusTypePendingCancel   OrderStatusType = "PENDING_CANCEL"
	OrderStatusTypeRejected        OrderStatusType = "REJECTED"
	OrderStatusTypeExpired         OrderStatusType = "EXPIRED"
	OrderStatusExpiredInMatch      OrderStatusType = "EXPIRED_IN_MATCH" // STP Expired
	OrderStatusTypeNewInsurance    OrderStatusType = "NEW_INSURANCE"
	OrderStatusTypeNewADL          OrderStatusType = "NEW_ADL"
)
