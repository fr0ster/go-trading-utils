package grid

import (
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/google/btree"
)

type (
	Record struct {
		Price     float64
		Quantity  float64
		OrderId   int64
		UpPrice   float64
		DownPrice float64
		OrderSide types.OrderSide
	}
)

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
	return g.Quantity
}

func (g *Record) SetQuantity(quantity float64) {
	g.Quantity = quantity
}

func (g *Record) GetOrderId() int64 {
	return g.OrderId
}

func (g *Record) SetOrderId(orderId int64) {
	g.OrderId = orderId
}

func (g *Record) GetUpPrice() float64 {
	return g.UpPrice
}

func (g *Record) SetUpPrice(upPrice float64) {
	g.UpPrice = upPrice
}

func (g *Record) GetDownPrice() float64 {
	return g.DownPrice
}

func (g *Record) SetDownPrice(downPrice float64) {
	g.DownPrice = downPrice
}

func (g *Record) GetOrderSide() types.OrderSide {
	return g.OrderSide
}

func (g *Record) SetOrderSide(orderSide types.OrderSide) {
	g.OrderSide = orderSide
}

func NewRecord(orderId int64, price float64, quantity float64, upPrice float64, downPrice float64, orderSide types.OrderSide) *Record {
	return &Record{
		Price:     price,
		Quantity:  quantity,
		OrderId:   orderId,
		UpPrice:   upPrice,
		DownPrice: downPrice,
		OrderSide: orderSide,
	}
}
