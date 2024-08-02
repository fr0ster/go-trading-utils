package orders_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/stretchr/testify/assert"

	spot_orders "github.com/fr0ster/go-trading-utils/binance/spot/orders"
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
	api_key := os.Getenv("SPOT_TEST_BINANCE_API_KEY")
	api_secret := os.Getenv("SPOT_TEST_BINANCE_SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, api_secret)
	orders := orders_types.New(
		symbol, // symbol
		spot_orders.UserDataStreamCreator(
			client,
			spot_orders.CallBackCreator(),
			spot_orders.WsErrorHandlerCreator()), // userDataStream
		spot_orders.CreateOrderCreator(
			client,
			stepSizeExp,
			tickSizeExp), // createOrder
		spot_orders.GetOpenOrdersCreator(client),   // getOpenOrders
		spot_orders.GetAllOrdersCreator(client),    // getAllOrders
		spot_orders.GetOrderCreator(client),        // getOrder
		spot_orders.CancelOrderCreator(client),     // cancelOrder
		spot_orders.CancelAllOrdersCreator(client), // cancelAllOrders
		quit) // quit
	assert.NotNil(t, orders)
	orders.StreamStart()
	orders.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
