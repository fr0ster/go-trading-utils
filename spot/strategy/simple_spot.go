package strategy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/orders"
	"github.com/fr0ster/go-binance-utils/spot/streams"
	"github.com/fr0ster/go-binance-utils/utils"
)

func SimpleSpot(client *binance.Client, symbolname, quantity, price, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta string) {
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error starting user stream: %v", err)
	}
	fmt.Println("ListenKey:", listenKey)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Price:", price)
	fmt.Println("StopPriceSL:", stopPriceSL)
	fmt.Println("PriceSL:", priceSL)
	fmt.Println("StopPriceTP:", stopPriceTP)
	fmt.Println("PriceTP:", priceTP)

	wsHandler, executeOrderChan := streams.GetFilledOrderHandler()

	_, _, err = streams.StartUserDataStream(listenKey, wsHandler, utils.HandleErr)
	if err != nil {
		log.Fatalf("Error serving user data websocket: %v", err)
	}

	order, err := orders.NewLimitOrder(
		client,
		symbolname,
		binance.SideTypeBuy,
		quantity,
		price,
		binance.TimeInForceTypeGTC)
	if err != nil {
		log.Fatalf("Error placing order: %v", err)
	}

	go func() {
		for event := range executeOrderChan {
			if event.OrderUpdate.Id == order.OrderID || event.OrderUpdate.ClientOrderId == order.ClientOrderID {
				fmt.Printf("Order executed: %v\n", order)
				handleOrders(
					client,                     // client,
					order,                      // order,
					binance.SideTypeSell,       // side,
					binance.TimeInForceTypeGTC, // timeInForce,
					priceSL,                    // priceSL,
					priceTP,                    // priceTP,
					stopPriceSL,                // stopPriceSL,
					stopPriceTP,                // stopPriceTP,
					trailingDelta)              // trailingDelta
				stop <- syscall.SIGSTOP
			}
		}
	}()

	utils.HandleShutdown(stop, 30*time.Second)
}

func handleOrders(
	client *binance.Client,
	order *binance.CreateOrderResponse,
	side binance.SideType,
	timeInForce binance.TimeInForceType,
	priceSL,
	priceTP,
	stopPriceSL,
	stopPriceTP,
	trailingDelta string) {
	stopLossOrder, err := orders.NewStopLossLimitOrder(
		client,
		order,
		order.Symbol,
		side,
		timeInForce,
		order.ExecutedQuantity,
		priceSL,
		stopPriceSL,
		trailingDelta)
	if err != nil {
		if order.Status == binance.OrderStatusTypeNew || order.Status == binance.OrderStatusTypePartiallyFilled {
			handleCancelOrder(client, order)
		}
		log.Fatalf("Error creating stop loss order: %v, StopPrice: %v, TrailingDelta: %v", err, stopPriceSL, trailingDelta)
	} else {
		fmt.Printf("StopLossOrder: %v\n", stopLossOrder)
	}

	takeProfitOrder, err := orders.NewTakeProfitLimitOrder(
		client,
		order,
		order.Symbol,
		side,
		timeInForce,
		order.ExecutedQuantity,
		priceTP,
		stopPriceTP,
		trailingDelta)
	if err != nil {
		if stopLossOrder.Status == binance.OrderStatusTypeNew || stopLossOrder.Status == binance.OrderStatusTypePartiallyFilled {
			handleCancelOrders(client, order, stopLossOrder)
		}
		log.Fatalf("Error creating take profit order: %v,  StopPrice: %v, TrailingDelta: %v", err, stopPriceTP, trailingDelta)
	} else {
		fmt.Printf("TakeProfitOrder: %v\n", takeProfitOrder)
	}
}

func handleCancelOrder(client *binance.Client, order *binance.CreateOrderResponse) {
	_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
	if err != nil {
		log.Fatalf("Error canceling order: %v", err)
	}
}

func handleCancelOrders(client *binance.Client, orders ...*binance.CreateOrderResponse) {
	for _, order := range orders {
		handleCancelOrder(client, order)
	}
}
