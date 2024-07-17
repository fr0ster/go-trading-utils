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

// Функції для btree.Btree
func (i *QuantityItem) Less(than btree.Item) bool {
	return i.quantity < than.(*QuantityItem).quantity
}

func (i *QuantityItem) Equal(than btree.Item) bool {
	return i.quantity == than.(*QuantityItem).quantity
}

// CRUD
func (i *QuantityItem) GetDepth(price PriceType) *DepthItem {
	if val := i.depths.Get(NewDepthItem(price)); val != nil {
		return val.(*DepthItem)
	} else {
		return nil
	}
}

func (i *QuantityItem) Add(price PriceType, quantity QuantityType) {
	if quantity == i.quantity {
		i.depths.ReplaceOrInsert(NewDepthItem(price, quantity))
	}
}

func (i *QuantityItem) Delete(price PriceType, quantity QuantityType) {
	if old := i.GetDepth(price); old != nil {
		i.depths.Delete(NewDepthItem(price))
	}
}

// Робота з Мінімальними та Максимальними значеннями
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

func (i *QuantityItem) IsEmpty() bool {
	return (i.depths.Len() == 0)
}

// Конструктори
func NewQuantityItem(price PriceType, quantity QuantityType, degree int) *QuantityItem {
	item := &QuantityItem{quantity: quantity, depths: btree.New(degree)}
	item.Add(price, quantity)
	return item
}
