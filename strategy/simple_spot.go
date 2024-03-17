package strategy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fr0ster/go-trading-utils/binance/spot"
	"github.com/fr0ster/go-trading-utils/interfaces"

	"github.com/fr0ster/go-trading-utils/utils"
)

func SimpleSpot(client *interfaces.Client, symbolname, quantity, price, stopPriceSL, priceSL, stopPriceTP, priceTP, trailingDelta string) {
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	listenKey, err :=
		spot.NewClient(apiKey, secretKey, true).
			GetClient().
			NewStartUserStreamService().
			Do(context.Background())
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

	// inChannel := make(chan *binance.WsUserDataEvent, 1)
	// _, _, err = streams.NewUserDataStream(listenKey).Start()
	// if err != nil {
	// 	log.Fatalf("Error starting user data stream: %v", err)
	// }
	// symbol := binance.SymbolType(symbolname)

	// executeOrderChan := streams.GetFilledOrdersGuard(inChannel)

	// order, err := orders.NewLimitOrder(
	// 	client,
	// 	symbol,
	// 	binance.SideTypeBuy,
	// 	quantity,
	// 	price,
	// 	binance.TimeInForceTypeGTC)
	// if err != nil {
	// 	log.Fatalf("Error placing order: %v", err)
	// }

	// go func() {
	// 	for event := range executeOrderChan {
	// 		if event.OrderUpdate.Id == order.OrderID || event.OrderUpdate.ClientOrderId == order.ClientOrderID {
	// 			fmt.Printf("Order executed: %v\n", order)
	// 			handleOrders(
	// 				client,                     // client,
	// 				order,                      // order,
	// 				binance.SideTypeSell,       // side,
	// 				binance.TimeInForceTypeGTC, // timeInForce,
	// 				priceSL,                    // priceSL,
	// 				priceTP,                    // priceTP,
	// 				stopPriceSL,                // stopPriceSL,
	// 				stopPriceTP,                // stopPriceTP,
	// 				trailingDelta)              // trailingDelta
	// 			stop <- syscall.SIGSTOP
	// 		}
	// 	}
	// }()

	utils.HandleShutdown(stop, 30*time.Second)
}

// func handleOrders(
// 	client *binance.Client,
// 	order *binance.CreateOrderResponse,
// 	side binance.SideType,
// 	timeInForce binance.TimeInForceType,
// 	priceSL,
// 	priceTP,
// 	stopPriceSL,
// 	stopPriceTP,
// 	trailingDelta string) {
// 	stopLossOrder, err := orders.NewStopLossLimitOrder(
// 		client,
// 		order,
// 		binance.SymbolType(order.Symbol), // Convert order.Symbol to binance.SymbolType
// 		side,
// 		timeInForce,
// 		order.ExecutedQuantity,
// 		priceSL,
// 		stopPriceSL,
// 		trailingDelta)
// 	if err != nil {
// 		if order.Status == binance.OrderStatusTypeNew || order.Status == binance.OrderStatusTypePartiallyFilled {
// 			handleCancelOrder(client, order)
// 		}
// 		log.Fatalf("Error creating stop loss order: %v, StopPrice: %v, TrailingDelta: %v", err, stopPriceSL, trailingDelta)
// 	} else {
// 		fmt.Printf("StopLossOrder: %v\n", stopLossOrder)
// 	}

// 	symbol := binance.SymbolType(order.Symbol)
// 	takeProfitOrder, err := orders.NewTakeProfitLimitOrder(
// 		client,
// 		order,
// 		symbol,
// 		side,
// 		timeInForce,
// 		order.ExecutedQuantity,
// 		priceTP,
// 		stopPriceTP,
// 		trailingDelta)
// 	if err != nil {
// 		if stopLossOrder.Status == binance.OrderStatusTypeNew || stopLossOrder.Status == binance.OrderStatusTypePartiallyFilled {
// 			handleCancelOrders(client, order, stopLossOrder)
// 		}
// 		log.Fatalf("Error creating take profit order: %v,  StopPrice: %v, TrailingDelta: %v", err, stopPriceTP, trailingDelta)
// 	} else {
// 		fmt.Printf("TakeProfitOrder: %v\n", takeProfitOrder)
// 	}
// }

// func handleCancelOrder(client *binance.Client, order *binance.CreateOrderResponse) {
// 	_, err := client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
// 	if err != nil {
// 		log.Fatalf("Error canceling order: %v", err)
// 	}
// }

// func handleCancelOrders(client *binance.Client, orders ...*binance.CreateOrderResponse) {
// 	for _, order := range orders {
// 		handleCancelOrder(client, order)
// 	}
// }
