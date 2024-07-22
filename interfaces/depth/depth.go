package depth

import (
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

type (
	Depth interface {
		Lock()
		Unlock()
		TryLock() bool
		GetAsks() *asks_types.Asks
		GetBids() *bids_types.Bids
		// AskAscend(iter func(btree.Item) bool)
		// AskDescend(iter func(btree.Item) bool)
		// BidAscend(iter func(btree.Item) bool)
		// BidDescend(iter func(btree.Item) bool)
		// GetAsk(price types.PriceType) btree.Item
		// GetBid(price types.PriceType) btree.Item
		// SetAsk(price types.PriceType, quantity types.QuantityType) error
		// SetBid(price types.PriceType, quantity types.QuantityType) error
		// ClearAsks()
		// ClearBids()
		// DeleteAsk(price types.PriceType)
		// DeleteBid(price types.PriceType)
		// RestrictAskUp(price types.PriceType)
		// RestrictBidUp(price types.PriceType)
		// RestrictAskDown(price types.PriceType)
		// RestrictBidDown(price types.PriceType)
		UpdateAsk(*items_types.Ask) bool
		UpdateBid(*items_types.Bid) bool
	}
)
