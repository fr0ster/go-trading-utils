package asks

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Asks) GetFilteredByPercent(f ...items_types.DepthFilter) (tree *btree.BTree, summa, max, min items_types.QuantityType) {
	return d.tree.GetFilteredByPercent(f...)
}

func (d *Asks) GetSummaByRange(first, last items_types.PriceType, f ...items_types.DepthFilter) (summa, max, min items_types.QuantityType) {
	return d.tree.GetSummaByRange(first, last, f...)
}
