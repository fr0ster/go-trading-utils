package depth

import (
	"github.com/google/btree"
)

type (
	Depth interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error)
		AskAscend(iter func(btree.Item) bool)
		AskDescend(iter func(btree.Item) bool)
		BidAscend(iter func(btree.Item) bool)
		BidDescend(iter func(btree.Item) bool)
		GetAsk(price float64) *DepthItemType
		GetBid(price float64) *DepthItemType
		SetAsk(price float64, quantity float64)
		SetBid(price float64, quantity float64)
		UpdateAsk(price float64, quantity float64)
		UpdateBid(price float64, quantity float64)
		GetMaxAsks() *DepthItemType
		GetMaxBids() *DepthItemType
		GetMinAsks() *DepthItemType
		GetMinBids() *DepthItemType
		GetBidLocalMaxima() *btree.BTree
		GetAskLocalMaxima() *btree.BTree
	}
	DepthItemType struct {
		Price    float64
		Quantity float64
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItemType).Price
}

func (i *DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItemType).Price
}
