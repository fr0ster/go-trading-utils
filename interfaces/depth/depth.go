package depth_interface

import (
	"sync"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	Depth interface {
		GetItem() *Depth
		Less(than interface{}) bool
		Equal(than interface{}) bool
	}
	Depths interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool) *Depths
		GetItem(price types.Price) (*DepthItemType, bool)
		SetItem(value DepthItemType)
		GetByPrices(minPrice, maxPrice types.Price) *DepthBTree
		GetMaxBids() *DepthBTree
		GetMaxAsks() *DepthBTree
		GetMaxBidQtyMaxAskQty() (maxBidNode *DepthItemType, maxAskNode *DepthItemType)
		GetMaxBidMinAsk() (maxBid *DepthItemType, minAsk *DepthItemType)
		GetBidQtyLocalMaxima() *DepthBTree
		GetAskQtyLocalMaxima() *DepthBTree
		Show()
	}
	DepthItemType struct {
		Price           types.Price
		AskLastUpdateID int64
		AskQuantity     types.Price
		BidLastUpdateID int64
		BidQuantity     types.Price
	}
	// Btree      btree.BTree
	// Mutex      sync.Mutex
	Degree     int
	DepthBTree struct {
		btree.BTree
		sync.Mutex
		Degree
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItemType).Price
}

func (i *DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItemType).Price
}
