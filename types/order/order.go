package order

import (
	types "github.com/fr0ster/go-trading-utils/types"
)

const (
	DepthSideAsk                   types.DepthSide       = "ASK"
	DepthSideBid                   types.DepthSide       = "BID"
	SideTypeBuy                    types.OrderSide       = "BUY"
	SideTypeSell                   types.OrderSide       = "SELL"
	OrderTypeLimit                 types.OrderType       = "LIMIT"
	OrderTypeMarket                types.OrderType       = "MARKET"
	OrderTypeLimitMaker            types.OrderType       = "LIMIT_MAKER"
	OrderTypeStopLoss              types.OrderType       = "STOP_LOSS"
	OrderTypeStopLossLimit         types.OrderType       = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit            types.OrderType       = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit       types.OrderType       = "TAKE_PROFIT_LIMIT"
	OrderStatusTypeNew             types.OrderStatusType = "NEW"
	OrderStatusTypePartiallyFilled types.OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusTypeFilled          types.OrderStatusType = "FILLED"
	OrderStatusTypeCanceled        types.OrderStatusType = "CANCELED"
	OrderStatusTypePendingCancel   types.OrderStatusType = "PENDING_CANCEL"
	OrderStatusTypeRejected        types.OrderStatusType = "REJECTED"
	OrderStatusTypeExpired         types.OrderStatusType = "EXPIRED"
	OrderStatusExpiredInMatch      types.OrderStatusType = "EXPIRED_IN_MATCH" // STP Expired
	OrderStatusTypeNewInsurance    types.OrderStatusType = "NEW_INSURANCE"
	OrderStatusTypeNewADL          types.OrderStatusType = "NEW_ADL"
)
