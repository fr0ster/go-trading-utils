package depth

import (
	"github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

type (
	Depth interface {
		Lock()
		Unlock()
		AskAscend(iter func(btree.Item) bool)
		AskDescend(iter func(btree.Item) bool)
		BidAscend(iter func(btree.Item) bool)
		BidDescend(iter func(btree.Item) bool)
		GetAsk(price types.PriceType) btree.Item
		GetBid(price types.PriceType) btree.Item
		SetAsk(price types.PriceType, quantity types.QuantityType) error
		SetBid(price types.PriceType, quantity types.QuantityType) error
		ClearAsks()
		ClearBids()
		DeleteAsk(price types.PriceType)
		DeleteBid(price types.PriceType)
		RestrictAskUp(price types.PriceType)
		RestrictBidUp(price types.PriceType)
		RestrictAskDown(price types.PriceType)
		RestrictBidDown(price types.PriceType)
		UpdateAsk(price types.PriceType, quantity types.QuantityType) bool
		UpdateBid(price types.PriceType, quantity types.QuantityType) bool
	}
)
