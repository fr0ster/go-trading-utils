package orders_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	futures_orders "github.com/fr0ster/go-trading-utils/binance/futures/orders"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

func TestEvents(t *testing.T) {
	t.Parallel()
	var (
		quit = make(chan struct{})
	)
	symbol := "BTCUSDT"
	stepSizeExp := 3
	tickSizeExp := 1
	t.Log("TestEvents")
	client := testutil.FuturesClient(t)
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
	orders.StreamStart()
	orders.ResetEvent(fmt.Errorf("test"))
	fmt.Println("test pass")
	time.Sleep(3 * time.Second)
	close(quit)
}
