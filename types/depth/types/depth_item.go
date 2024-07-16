package types

import (
	"github.com/google/btree"
)

type (
	DepthItem struct {
		price    PriceType
		quantity QuantityType
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

func (i *DepthItem) GetPrice() PriceType {
	return i.price
}

func (i *DepthItem) SetPrice(price PriceType) {
	i.price = price
}

func (i *DepthItem) GetQuantity() QuantityType {
	return i.quantity
}

func (i *DepthItem) SetQuantity(quantity QuantityType) {
	i.quantity = quantity
}

// GetAskDeviation implements depth_interface.Depths.
func (d *DepthItem) GetQuantityDeviation(middle QuantityType) float64 {
	return float64(d.quantity - middle)
}

func NewDepthItem(price PriceType, quantity ...QuantityType) *DepthItem {
	if len(quantity) > 0 {
		return &DepthItem{price: price, quantity: quantity[0]}
	} else {
		return &DepthItem{price: price}
	}
}
