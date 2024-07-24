package types

type (
	OrderSide        string
	OrderType        string
	SideType         string
	TimeInForceType  string
	QuantityType     string
	OrderStatusType  string
	WorkingType      string
	PositionSideType string
	DepthSide        string

	StreamFunction       func() (chan struct{}, chan struct{}, error)
	InitFunction         func() (err error)
	ErrorHandlerFunction func(err error)

	AccountType     string
	MarginType      string
	ProgressionType string
	StageType       string
	StrategyType    string
)

const (
	DepthSideAsk DepthSide = "ASK"
	DepthSideBid DepthSide = "BID"
	SideTypeBuy  OrderSide = "BUY"
	SideTypeSell OrderSide = "SELL"
	SideTypeNone OrderSide = "NONE"
)
