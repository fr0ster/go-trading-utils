package grid

import (
	"github.com/fr0ster/go-trading-utils/types"
	depth_items "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

type (
	OrderIdType int64
	Record      struct {
		Price     depth_items.PriceType
		quantity  depth_items.QuantityType
		orderId   int64
		uPrice    depth_items.PriceType
		downPrice depth_items.PriceType
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

func (g *Record) GetPrice() depth_items.PriceType {
	return g.Price
}

func (g *Record) SetPrice(price depth_items.PriceType) {
	g.Price = price
}

func (g *Record) GetQuantity() depth_items.QuantityType {
	return g.quantity
}

func (g *Record) SetQuantity(quantity depth_items.QuantityType) {
	g.quantity = quantity
}

func (g *Record) GetOrderId() int64 {
	return g.orderId
}

func (g *Record) SetOrderId(orderId int64) {
	g.orderId = orderId
}

func (g *Record) GetUpPrice() depth_items.PriceType {
	return g.uPrice
}

func (g *Record) SetUpPrice(upPrice depth_items.PriceType) {
	g.uPrice = upPrice
}

func (g *Record) GetDownPrice() depth_items.PriceType {
	return g.downPrice
}

func (g *Record) SetDownPrice(downPrice depth_items.PriceType) {
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
	price depth_items.PriceType,
	quantity depth_items.QuantityType,
	upPrice depth_items.PriceType,
	downPrice depth_items.PriceType,
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
