package types

import (
	"github.com/google/btree"
)

type (
	NormalizedItem struct {
		Price     float64
		MinMax    *btree.BTree
		DepthItem *btree.BTree
	}
)

func (i *NormalizedItem) Less(than btree.Item) bool {
	return i.Price < than.(*NormalizedItem).Price
}

func (i *NormalizedItem) Equal(than btree.Item) bool {
	return i.Price == than.(*NormalizedItem).Price
}

func (i *NormalizedItem) GetDepth(price float64) (depthItem *DepthItem) {
	if i.DepthItem != nil {
		depthItem = i.DepthItem.Get(&DepthItem{Price: price}).(*DepthItem)
	}
	return
}

func (i *NormalizedItem) SetDepth(depth *DepthItem) {
	i.DepthItem.ReplaceOrInsert(depth)
}

func (i *NormalizedItem) DeleteDepth(depth *DepthItem) {
	i.DepthItem.Delete(depth)
}

func (i *NormalizedItem) GetMinMax(quantity float64) (minMax *QuantityItem) {
	if i.MinMax != nil {
		if val := i.MinMax.Get(&QuantityItem{Quantity: quantity}); val != nil {
			minMax = val.(*QuantityItem)
		}
	}
	return
}

func (i *NormalizedItem) SetMinMax(minMax *QuantityItem) {
	i.MinMax.ReplaceOrInsert(minMax)
}

func (i *NormalizedItem) DeleteMinMax(minMax *QuantityItem) {
	i.MinMax.Delete(minMax)
}
