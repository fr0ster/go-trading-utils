package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func NewAsks(
	degree int,
	symbol string,
	targetPercent float64,
	limitDepth depths_types.DepthAPILimit,
	expBase int,
	rate ...depths_types.DepthStreamRate) *Asks {
	return &Asks{tree: depths_types.New(degree, symbol, targetPercent, limitDepth, expBase, rate...)}
}

// Get implements depth_interface.Depths.
func (d *Asks) GetTree() *btree.BTree {
	return d.tree.GetTree()
}

// Set implements depth_interface.Depths.
func (d *Asks) SetTree(tree *btree.BTree) {
	d.tree.SetTree(tree)
}

func (a *Asks) Get(item *items_types.Ask) *items_types.Ask {
	if val := a.tree.Get((*items_types.DepthItem)(item)); val != nil {
		return (*items_types.Ask)(val)
	} else {
		return nil
	}
}

func (a *Asks) Set(item *items_types.Ask) (err error) {
	return a.tree.Set((*items_types.DepthItem)(item))
}

func (a *Asks) Delete(item *items_types.Ask) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Asks) Update(item *items_types.Ask) bool {
	return a.tree.Update((*items_types.DepthItem)(item))
}

// Обертки вокруг методів з Depths
func (d *Asks) Count() int {
	return d.tree.Count()
}

func (d *Asks) GetSummaQuantity() items_types.QuantityType {
	return d.tree.GetSummaQuantity()
}

func (d *Asks) GetSummaValue() items_types.ValueType {
	return d.tree.GetSummaValue()
}
