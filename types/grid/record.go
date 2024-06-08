package grid

import (
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	OrderIdType int64
	Record      struct {
		Price     float64
		quantity  float64
		orderId   int64
		uPrice    float64
		downPrice float64
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

func (g *Record) GetPrice() float64 {
	return g.Price
}

func (g *Record) SetPrice(price float64) {
	g.Price = price
}

func (g *Record) GetQuantity() float64 {
	return g.quantity
}

func (g *Record) SetQuantity(quantity float64) {
	g.quantity = quantity
}

func (g *Record) GetOrderId() int64 {
	return g.orderId
}

func (g *Record) SetOrderId(orderId int64) {
	g.orderId = orderId
}

func (g *Record) GetUpPrice() float64 {
	return g.uPrice
}

func (g *Record) SetUpPrice(upPrice float64) {
	g.uPrice = upPrice
}

func (g *Record) GetDownPrice() float64 {
	return g.downPrice
}

func (g *Record) SetDownPrice(downPrice float64) {
	g.downPrice = downPrice
}

func (g *Record) GetOrderSide() types.OrderSide {
	return g.orderSide
}

func (g *Record) SetOrderSide(orderSide types.OrderSide) {
	g.orderSide = orderSide
}

func NewRecord(orderId int64, price float64, quantity float64, upPrice float64, downPrice float64, orderSide types.OrderSide) *Record {
	return &Record{
		Price:     price,
		quantity:  quantity,
		orderId:   orderId,
		uPrice:    upPrice,
		downPrice: downPrice,
		orderSide: orderSide,
	}
}
