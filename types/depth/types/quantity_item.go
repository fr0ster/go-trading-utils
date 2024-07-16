package types

import (
	"github.com/google/btree"
)

type (
	QuantityItem struct {
		quantity QuantityType
		depths   *btree.BTree
	}
)

func (i *QuantityItem) Less(than btree.Item) bool {
	return i.quantity < than.(*QuantityItem).quantity
}

func (i *QuantityItem) Equal(than btree.Item) bool {
	return i.quantity == than.(*QuantityItem).quantity
}

func (i *QuantityItem) GetDepths() *btree.BTree {
	return i.depths
}

func (i *QuantityItem) GetDepth(price PriceType) *DepthItem {
	if val := i.depths.Get(NewDepthItem(price)); val != nil {
		return val.(*DepthItem)
	} else {
		return nil
	}
}

func (i *QuantityItem) GetDepthMin() *DepthItem {
	if val := i.depths.Min(); val != nil {
		return val.(*DepthItem)
	} else {
		return nil
	}
}

func (i *QuantityItem) GetDepthMax() *DepthItem {
	if val := i.depths.Max(); val != nil {
		return val.(*DepthItem)
	} else {
		return nil
	}
}

func (i *QuantityItem) SetDepth(depth *DepthItem) {
	i.depths.ReplaceOrInsert(depth)
}

func (i *QuantityItem) DeleteDepth(depth *DepthItem) {
	i.depths.Delete(depth)
}

func NewQuantityItem(price PriceType, quantity QuantityType, degree int) *QuantityItem {
	item := &QuantityItem{quantity: quantity, depths: btree.New(degree)}
	item.SetDepth(NewDepthItem(price, quantity))
	return item
}
