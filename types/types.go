package types

import "github.com/google/btree"

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

	OrderIdType int64
)

const (
	DepthSideAsk DepthSide = "ASK"
	DepthSideBid DepthSide = "BID"
	SideTypeBuy  OrderSide = "BUY"
	SideTypeSell OrderSide = "SELL"
	SideTypeNone OrderSide = "NONE"
)

// Функції для btree.Btree
func (i OrderIdType) Less(than btree.Item) bool {
	return i < than.(OrderIdType)
}

func (i OrderIdType) Equal(than btree.Item) bool {
	return i == than.(OrderIdType)
}
