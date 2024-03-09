package orders

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

func NewOrder(
	client *binance.Client,
	symbol string,
	orderType binance.OrderType,
	side binance.SideType,
	quantity,
	price string,
	timeInForce binance.TimeInForceType) (*binance.CreateOrderResponse, error) {
	order := client.NewCreateOrderService().
		Symbol(symbol).
		Type(orderType).
		Side(side).
		Quantity(quantity)
	if orderType == binance.OrderTypeLimit {
		order = order.Price(price)
	}
	return order.TimeInForce(timeInForce).Do(context.Background())
}
