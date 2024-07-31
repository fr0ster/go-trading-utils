package orders_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/stretchr/testify/assert"

	futures_orders "github.com/fr0ster/go-trading-utils/binance/futures/orders"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

func TestEvents(t *testing.T) {
	var (
		quit = make(chan struct{})
	)
	symbol := "BTCUSDT"
	stepSizeExp := 3
	tickSizeExp := 1
	t.Log("TestEvents")
	api_key := os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	api_secret := os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	futures.UseTestnet = true
	client := futures.NewClient(api_key, api_secret)
	orders := orders_types.New(
		symbol, // symbol
		futures_orders.UserDataStreamCreator(
			client,
			futures_orders.CallBackCreator(),
			futures_orders.WsErrorHandlerCreator()), // userDataStream
		futures_orders.CreateOrderCreator(
			client,
			stepSizeExp,
			tickSizeExp), // createOrder
		futures_orders.GetOpenOrdersCreator(client),   // getOpenOrders
		futures_orders.GetAllOrdersCreator(client),    // getAllOrders
		futures_orders.GetOrderCreator(client),        // getOrder
		futures_orders.CancelOrderCreator(client),     // cancelOrder
		futures_orders.CancelAllOrdersCreator(client), // cancelAllOrders
		quit) // quit
	assert.NotNil(t, orders)
	orders.UserDataEventStart()
	orders.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
