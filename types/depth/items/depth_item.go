package types

import (
	"github.com/google/btree"
)

type (
	Ask       DepthItem
	Bid       DepthItem
	DepthItem struct {
		price    PriceType
		quantity QuantityType
	}
	DepthFilter func(*DepthItem) bool
	DepthTester func(result *DepthItem, target *DepthItem) bool
)

// Функції для btree.Btree
func (i *DepthItem) Less(than btree.Item) bool {
	return i.price < than.(*DepthItem).price
}

func (i *DepthItem) Equal(than btree.Item) bool {
	return i.price == than.(*DepthItem).price
}
func (i *Ask) Less(than btree.Item) bool {
	return i.price < than.(*Ask).price
}

func (i *Ask) Equal(than btree.Item) bool {
	return i.price == than.(*Ask).price
}

func (i *Ask) GetDepthItem() *DepthItem {
	if i != nil {
		return (*DepthItem)(i)
	} else {
		return nil
	}
}

func (i *Bid) Less(than btree.Item) bool {
	return i.price < than.(*Bid).price
}

func (i *Bid) Equal(than btree.Item) bool {
	return i.price == than.(*Bid).price
}

func (i *Bid) GetDepthItem() *DepthItem {
	if i != nil {
		return (*DepthItem)(i)
	} else {
		return nil
	}
}

// CRUD
func (i *DepthItem) GetPrice() PriceType {
	if i != nil {
		return i.price
	} else {
		return 0
	}
}

func (i *DepthItem) SetPrice(price PriceType) {
	if i != nil {
		i.price = price
	}
}

func (i *DepthItem) GetQuantity() QuantityType {
	if i != nil {
		return i.quantity
	} else {
		return 0
	}
}

func (i *DepthItem) GetValue() ValueType {
	if i != nil {
		return ValueType(i.quantity) * ValueType(i.price)
	} else {
		return 0
	}
}

func (i *DepthItem) SetQuantity(quantity QuantityType) {
	if i != nil {
		i.quantity = quantity
	}
}

// Статистичні функції
func (d *DepthItem) GetQuantityDeviation(middle QuantityType) float64 {
	return float64(d.quantity - middle)
}

// Конструктори
func New(price PriceType, quantity ...QuantityType) *DepthItem {
	if len(quantity) > 0 {
		return &DepthItem{price: price, quantity: quantity[0]}
	} else {
		return &DepthItem{price: price}
	}
}

func NewAsk(price PriceType, quantity ...QuantityType) *Ask {
	if len(quantity) > 0 {
		return &Ask{price: price, quantity: quantity[0]}
	} else {
		return &Ask{price: price}
	}
}

func NewBid(price PriceType, quantity ...QuantityType) *Bid {
	if len(quantity) > 0 {
		return &Bid{price: price, quantity: quantity[0]}
	} else {
		return &Bid{price: price}
	}
}
