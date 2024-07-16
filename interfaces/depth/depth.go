package depth

import (
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
		GetAsk(price float64) btree.Item
		GetBid(price float64) btree.Item
		SetAsk(price float64, quantity float64) error
		SetBid(price float64, quantity float64) error
		ClearAsks()
		ClearBids()
		DeleteAsk(price float64)
		DeleteBid(price float64)
		RestrictAskUp(price float64)
		RestrictBidUp(price float64)
		RestrictAskDown(price float64)
		RestrictBidDown(price float64)
		UpdateAsk(price float64, quantity float64) bool
		UpdateBid(price float64, quantity float64) bool
	}
)
