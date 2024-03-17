package interfaces

import (
	"github.com/fr0ster/go-trading-utils/types"
)

type (
	DepthItemType struct {
		Price           types.Price
		AskLastUpdateID int64
		AskQuantity     types.Price
		BidLastUpdateID int64
		BidQuantity     types.Price
	}
	Depth interface {
		GetDepths(symbol string, limit int) *Depth
	}
)
