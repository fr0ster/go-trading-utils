package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func NewAsks(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth DepthAPILimit,
	expBase int,
	rate ...DepthStreamRate) *Asks {
	return &Asks{tree: New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}

func (a *Asks) GetDepths() *Depths {
	return a.tree
}

// Get implements depth_interface.Depths.
func (d *Asks) GetTree() *btree.BTree {
	return d.tree.GetTree()
}

// Set implements depth_interface.Depths.
func (d *Asks) SetTree(tree *btree.BTree) {
	d.tree.SetTree(tree)
}

func (a *Asks) Get(item *types.Ask) *types.Ask {
	if val := a.tree.Get((*types.DepthItem)(item)); val != nil {
		return (*types.Ask)(val)
	} else {
		return nil
	}
}

func (a *Asks) Set(item *types.Ask) (err error) {
	return a.tree.Set((*types.DepthItem)(item))
}

func (a *Asks) Delete(item *types.Ask) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Asks) Update(item *types.Ask) bool {
	return a.tree.Update((*types.DepthItem)(item))
}

// func (a *Asks) Count() int {
// 	return a.tree.Count()
// }

// func (a *Asks) Symbol() string {
// 	return a.tree.Symbol()
// }

// func (a *Asks) GetSummaQuantity() types.QuantityType {
// 	return a.tree.GetSummaQuantity()
// }
