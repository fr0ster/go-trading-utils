package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func NewBids(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth depths_types.DepthAPILimit,
	expBase int,
	rate ...depths_types.DepthStreamRate) *Bids {
	return &Bids{tree: depths_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}

func (a *Bids) GetDepths() *depths_types.Depths {
	return a.tree
}

// Get implements depth_interface.Depths.
func (d *Bids) GetTree() *btree.BTree {
	return d.tree.GetTree()
}

// Set implements depth_interface.Depths.
func (d *Bids) SetTree(tree *btree.BTree) {
	d.tree.SetTree(tree)
}

func (a *Bids) Get(item *items_types.Bid) *items_types.Bid {
	if val := a.tree.Get((*items_types.DepthItem)(item)); val != nil {
		return (*items_types.Bid)(val)
	} else {
		return nil
	}
}

func (a *Bids) Set(item *items_types.Bid) (err error) {
	return a.tree.Set((*items_types.DepthItem)(item))
}

func (a *Bids) Delete(item *items_types.Bid) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Bids) Update(item *items_types.Bid) bool {
	return a.tree.Update((*items_types.DepthItem)(item))
}

// Обертки вокруг методів з Depths

func (d *Bids) Count() int {
	return d.tree.Count()
}

func (d *Bids) GetSummaQuantity() items_types.QuantityType {
	return d.tree.GetSummaQuantity()
}

func (d *Bids) GetSummaValue() items_types.ValueType {
	return d.tree.GetSummaValue()
}
