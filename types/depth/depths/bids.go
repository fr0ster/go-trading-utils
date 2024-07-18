package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func NewBids(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth DepthAPILimit,
	expBase int,
	rate ...DepthStreamRate) *Bids {
	return &Bids{tree: New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}

func (a *Bids) GetDepths() *Depths {
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

func (a *Bids) Get(item *types.Bid) *types.Bid {
	if val := a.tree.Get((*types.DepthItem)(item)); val != nil {
		return (*types.Bid)(val)
	} else {
		return nil
	}
}

func (a *Bids) Set(item *types.Bid) (err error) {
	return a.tree.Set((*types.DepthItem)(item))
}

func (a *Bids) Delete(item *types.Bid) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Bids) Update(item *types.Bid) bool {
	return a.tree.Update((*types.DepthItem)(item))
}
