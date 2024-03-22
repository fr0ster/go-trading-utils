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
		SetAsk(price float64, quantity float64)
		SetBid(price float64, quantity float64)
		UpdateAsk(price float64, quantity float64)
		UpdateBid(price float64, quantity float64)
	}
)
