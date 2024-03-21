package strategy

import (
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/google/btree"
)

type (
	DepthItemType struct {
		Price    float64
		Quantity float64
	}
	Depths struct {
		ds depth_interface.Depth
	}
)

// GetMaxAsks implements depth_interface.Depths.
func (d *Depth) GetMaxAsks() *DepthItemType {
	item := d.asks.Max()
	return item.(*DepthItemType)
}

// GetMaxBids implements depth_interface.Depths.
func (d *Depth) GetMaxBids() *DepthItemType {
	item := d.bids.Max()
	if item == nil {
		return nil
	}
	return item.(*DepthItemType)
}

// GetMinAsks implements depth_interface.Depths.
func (d *Depth) GetMinAsks() *DepthItemType {
	item := d.asks.Min()
	return item.(*DepthItemType)
}

// GetMinBids implements depth_interface.Depths.
func (d *Depth) GetMinBids() *DepthItemType {
	item := d.bids.Min()
	if item == nil {
		return nil
	}
	return item.(*DepthItemType)
}

// GetBidLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetBidLocalMaxima() *btree.BTree {
	maximaTree := btree.New(d.degree)
	var prev, current, next *DepthItemType
	d.BidAscend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
			(current != nil && prev == nil && current.Quantity > next.Quantity) {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

// GetAskLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetAskLocalMaxima() *btree.BTree {
	maximaTree := btree.New(d.degree)
	var prev, current, next *DepthItemType
	d.AskAscend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
			(current != nil && prev == nil && current.Quantity > next.Quantity) {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}
