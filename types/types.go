package types

type (
	OrderSide            string
	DepthSide            string
	StreamFunction       func() (chan struct{}, chan struct{}, error)
	InitFunction         func() (err error)
	ErrorHandlerFunction func(err error)
)

const (
	DepthSideAsk DepthSide = "ASK"
	DepthSideBid DepthSide = "BID"
	SideTypeBuy  OrderSide = "BUY"
	SideTypeSell OrderSide = "SELL"
	SideTypeNone OrderSide = "NONE"
)
