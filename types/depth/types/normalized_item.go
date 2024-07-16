package types

import (
	"github.com/google/btree"
)

type (
	NormalizedItem struct {
		degree   int
		price    float64
		quantity float64
		minMax   *btree.BTree
		depths   *btree.BTree
	}
)

func (i *NormalizedItem) Less(than btree.Item) bool {
	return i.price < than.(*NormalizedItem).price
}

func (i *NormalizedItem) Equal(than btree.Item) bool {
	return i.price == than.(*NormalizedItem).price
}

func (i *NormalizedItem) GetPrice() float64 {
	return i.price
}

func (i *NormalizedItem) GetQuantity() float64 {
	return i.quantity
}

func (i *NormalizedItem) SetQuantity(quantity float64) {
	i.quantity = quantity
}

func (i *NormalizedItem) GetDepth(price float64) (depthItem *DepthItem) {
	if i.depths != nil {
		depthItem = i.depths.Get(NewDepthItem(price)).(*DepthItem)
	}
	return
}

func (i *NormalizedItem) SetDepth(depth *DepthItem) {
	i.depths.ReplaceOrInsert(depth)
}

func (i *NormalizedItem) DeleteDepth(depth *DepthItem) {
	i.depths.Delete(depth)
}

func (i *NormalizedItem) GetMinMax(quantity float64) (minMax *QuantityItem) {
	if i.minMax != nil {
		if val := i.minMax.Get(&QuantityItem{quantity: quantity}); val != nil {
			minMax = val.(*QuantityItem)
		}
	}
	return
}

func (i *NormalizedItem) SetMinMax(minMax *QuantityItem) {
	i.minMax.ReplaceOrInsert(minMax)
}

func (i *NormalizedItem) DeleteMinMax(minMax *QuantityItem) {
	i.minMax.Delete(minMax)
}

func NewNormalizedItem(price float64, quantity float64, degree int) *NormalizedItem {
	item := &NormalizedItem{degree: degree, price: price, quantity: quantity, minMax: btree.New(degree), depths: btree.New(degree)}
	item.SetDepth(NewDepthItem(price, quantity))
	item.SetMinMax(NewQuantityItem(price, quantity, degree))
	return item
}
