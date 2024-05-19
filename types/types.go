package types

type (
	OrderSide string
	DepthSide string
)

const (
	DepthSideAsk DepthSide = "ASK"
	DepthSideBid DepthSide = "BID"
	SideTypeBuy  OrderSide = "BUY"
	SideTypeSell OrderSide = "SELL"
	SideTypeNone OrderSide = "NONE"
)
