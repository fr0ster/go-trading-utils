package types

import (
	"github.com/google/btree"
)

type (
	QuantityItem struct {
		Quantity float64
		Depths   *btree.BTree
	}
)

func (i *QuantityItem) Less(than btree.Item) bool {
	return i.Quantity < than.(*QuantityItem).Quantity
}

func (i *QuantityItem) Equal(than btree.Item) bool {
	return i.Quantity == than.(*QuantityItem).Quantity
}

func (i *QuantityItem) SetDepth(depth *DepthItem) {
	i.Depths.ReplaceOrInsert(depth)
}

func (i *QuantityItem) DeleteDepth(depth *DepthItem) {
	i.Depths.Delete(depth)
}
