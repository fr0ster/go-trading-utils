package asks

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Asks) GetFilteredByPercent(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min types.QuantityType) {
	return d.tree.GetFilteredByPercent(f...)
}

func (d *Asks) GetSummaByRange(first, last types.PriceType, f ...types.DepthFilter) (summa, max, min types.QuantityType) {
	return d.tree.GetSummaByRange(first, last, f...)
}

// RestrictUp implements depth_interface.Depths.
func (d *Asks) RestrictUp(price types.PriceType) {
	d.tree.RestrictUp(price)
}

// RestrictDown implements depth_interface.Depths.
func (d *Asks) RestrictDown(price types.PriceType) {
	d.tree.RestrictDown(price)
}
