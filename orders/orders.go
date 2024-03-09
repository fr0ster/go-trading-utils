package orders

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

func NewLimitOrder(
	client *binance.Client,
	symbol string,
	side binance.SideType,
	quantity,
	quoteOrderQty,
	price string,
	timeInForce binance.TimeInForceType) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Type(binance.OrderTypeLimit).
		Side(side).
		Quantity(quantity).
		Price(price).
		TimeInForce(timeInForce).Do(context.Background())
}

func NewMarketOrder(
	client *binance.Client,
	symbol string,
	side binance.SideType,
	quantity,
	quoteOrderQty,
	price string,
	timeInForce binance.TimeInForceType) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Type(binance.OrderTypeLimit).
		Side(side).
		Quantity(quantity).
		QuoteOrderQty(quoteOrderQty).
		Price(price).
		TimeInForce(timeInForce).Do(context.Background())
}

func NewLimitMakerOrder(
	client *binance.Client,
	symbol string,
	side binance.SideType,
	quantity,
	quoteOrderQty,
	price string,
	timeInForce binance.TimeInForceType) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Type(binance.OrderTypeLimit).
		Side(side).
		Quantity(quantity).
		QuoteOrderQty(quoteOrderQty).
		Price(price).
		TimeInForce(timeInForce).Do(context.Background())
}

func NewStopLossOrder(
	client *binance.Client,
	order *binance.CreateOrderResponse,
	symbol string,
	side binance.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(binance.OrderTypeStopLoss).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		TrailingDelta(trailingDelta).
		NewClientOrderID(order.ClientOrderID + "SL").
		Do(context.Background())
}

func NewTakeProfitOrder(
	client *binance.Client,
	order *binance.CreateOrderResponse,
	symbol string,
	side binance.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(binance.OrderTypeTakeProfit).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		TrailingDelta(trailingDelta).
		NewClientOrderID(order.ClientOrderID + "TP").
		Do(context.Background())
}

func NewStopLossLimitOrder(
	client *binance.Client,
	order *binance.CreateOrderResponse,
	symbol string,
	side binance.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(binance.OrderTypeStopLossLimit).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		TrailingDelta(trailingDelta).
		NewClientOrderID(order.ClientOrderID + "SL").
		Do(context.Background())
}

func NewTakeProfitLimitOrder(
	client *binance.Client,
	order *binance.CreateOrderResponse,
	symbol string,
	side binance.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*binance.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(binance.OrderTypeTakeProfitLimit).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		TrailingDelta(trailingDelta).
		NewClientOrderID(order.ClientOrderID + "TP").
		Do(context.Background())
}
