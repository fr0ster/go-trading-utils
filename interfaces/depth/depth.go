package depth

import (
	"github.com/google/btree"
)

type (
	Depths interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error)
		Ascend(iter func(btree.Item) bool)
		Descend(iter func(btree.Item) bool)
		GetItem(price float64) *DepthItemType
		SetItem(value DepthItemType)
		GetMaxBids() *DepthItemType
		GetMaxAsks() *DepthItemType
		GetMaxBidQtyMaxAskQty() (maxBidNode *DepthItemType, maxAskNode *DepthItemType)
		GetMaxBidMinAsk() (maxBid *DepthItemType, minAsk *DepthItemType)
		GetBidQtyLocalMaxima() *btree.BTree
		GetAskQtyLocalMaxima() *btree.BTree
		Show()
	}
	DepthItemType struct {
		Price       float64
		AskQuantity float64
		BidQuantity float64
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItemType).Price
}

func (i *DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItemType).Price
}
