package orders

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
)

func NewLimitOrder(
	client *futures.Client,
	symbol futures.SymbolType,
	side futures.SideType,
	quantity,
	price string,
	timeInForce futures.TimeInForceType) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Type(futures.OrderTypeLimit).
		Side(side).
		Quantity(quantity).
		Price(price).
		TimeInForce(timeInForce).Do(context.Background())
}

func NewMarketOrder(
	client *futures.Client,
	symbol futures.SymbolType,
	side futures.SideType,
	quantity string) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Type(futures.OrderTypeMarket).
		Side(side).
		Quantity(quantity).
		Do(context.Background())
}

func NewLimitMakerOrder(
	client *futures.Client,
	symbol futures.SymbolType,
	side futures.SideType,
	quantity,
	quoteOrderQty,
	price string,
	timeInForce futures.TimeInForceType) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Type(futures.OrderTypeLimit).
		Side(side).
		Quantity(quantity).
		Price(price).
		TimeInForce(timeInForce).Do(context.Background())
}

func NewStopOrder(
	client *futures.Client,
	order *futures.CreateOrderResponse,
	symbol futures.SymbolType,
	side futures.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Side(side).
		Type(futures.OrderTypeStop).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		NewClientOrderID(order.ClientOrderID + "SL").
		Do(context.Background())
}

func NewTakeProfitOrder(
	client *futures.Client,
	order *futures.CreateOrderResponse,
	symbol futures.SymbolType,
	side futures.SideType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Side(side).
		Type(futures.OrderTypeTakeProfit).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		NewClientOrderID(order.ClientOrderID + "TP").
		Do(context.Background())
}

func NewStopMarketOrder(
	client *futures.Client,
	order *futures.CreateOrderResponse,
	symbol futures.SymbolType,
	side futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)).
		Side(side).
		TimeInForce(timeInForce).
		Type(futures.OrderTypeStopMarket).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		NewClientOrderID(order.ClientOrderID + "SL").
		Do(context.Background())
}

func NewTakeProfitMarketOrder(
	client *futures.Client,
	order *futures.CreateOrderResponse,
	symbol futures.SymbolType,
	side futures.SideType,
	timeInForce futures.TimeInForceType,
	quantity,
	price,
	stopPrice,
	trailingDelta string) (*futures.CreateOrderResponse, error) {
	return client.NewCreateOrderService().
		Symbol(string(symbol)). // Convert symbol to string
		Side(side).
		TimeInForce(timeInForce).
		Type(futures.OrderTypeTakeProfitMarket).
		Quantity(quantity).
		Price(price).
		StopPrice(stopPrice).
		NewClientOrderID(order.ClientOrderID + "TP").
		Do(context.Background())
}

func CancelOrder(client *futures.Client, order *futures.CreateOrderResponse) (*futures.CancelOrderResponse, error) {
	return client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
}
