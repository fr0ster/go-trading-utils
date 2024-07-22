package grid

import (
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/google/btree"
)

type (
	OrderIdType int64
	Record      struct {
		Price     items_types.PriceType
		quantity  items_types.QuantityType
		orderId   int64
		uPrice    items_types.PriceType
		downPrice items_types.PriceType
		orderSide types.OrderSide
	}
)

func (g OrderIdType) Less(other btree.Item) bool {
	return g < other.(OrderIdType)
}

func (g OrderIdType) Equals(other btree.Item) bool {
	return g == other.(OrderIdType)
}

func (g *Record) Less(other btree.Item) bool {
	return (g.Price != 0 && g.Price < other.(*Record).Price)
}

func (g *Record) Equals(other btree.Item) bool {
	return (g.Price == other.(*Record).Price)
}

func (g *Record) GetPrice() items_types.PriceType {
	return g.Price
}

func (g *Record) SetPrice(price items_types.PriceType) {
	g.Price = price
}

func (g *Record) GetQuantity() items_types.QuantityType {
	return g.quantity
}

func (g *Record) SetQuantity(quantity items_types.QuantityType) {
	g.quantity = quantity
}

func (g *Record) GetOrderId() int64 {
	return g.orderId
}

func (g *Record) SetOrderId(orderId int64) {
	g.orderId = orderId
}

func (g *Record) GetUpPrice() items_types.PriceType {
	return g.uPrice
}

func (g *Record) SetUpPrice(upPrice items_types.PriceType) {
	g.uPrice = upPrice
}

func (g *Record) GetDownPrice() items_types.PriceType {
	return g.downPrice
}

func (g *Record) SetDownPrice(downPrice items_types.PriceType) {
	g.downPrice = downPrice
}

func (g *Record) GetOrderSide() types.OrderSide {
	return g.orderSide
}

func (g *Record) SetOrderSide(orderSide types.OrderSide) {
	g.orderSide = orderSide
}

func NewRecord(
	orderId int64,
	price items_types.PriceType,
	quantity items_types.QuantityType,
	upPrice items_types.PriceType,
	downPrice items_types.PriceType,
	orderSide types.OrderSide) *Record {
	return &Record{
		Price:     price,
		quantity:  quantity,
		orderId:   orderId,
		uPrice:    upPrice,
		downPrice: downPrice,
		orderSide: orderSide,
	}
}
