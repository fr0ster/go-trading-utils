package types

import (
	"github.com/google/btree"
)

type (
	DepthItem struct {
		price    float64
		quantity float64
	}
	DepthFilter func(*DepthItem) bool
	DepthTester func(result *DepthItem, target *DepthItem) bool
)

func (i *DepthItem) Less(than btree.Item) bool {
	return i.price < than.(*DepthItem).price
}

func (i *DepthItem) Equal(than btree.Item) bool {
	return i.price == than.(*DepthItem).price
}

func (i *DepthItem) GetPrice() float64 {
	return i.price
}

func (i *DepthItem) SetPrice(price float64) {
	i.price = price
}

func (i *DepthItem) GetQuantity() float64 {
	return i.quantity
}

func (i *DepthItem) SetQuantity(quantity float64) {
	i.quantity = quantity
}

// GetAskDeviation implements depth_interface.Depths.
func (d *DepthItem) GetQuantityDeviation(middle float64) float64 {
	return d.quantity - middle
}

func NewDepthItem(price float64, quantity ...float64) *DepthItem {
	if len(quantity) > 0 {
		return &DepthItem{price: price, quantity: quantity[0]}
	} else {
		return &DepthItem{price: price}
	}
}
