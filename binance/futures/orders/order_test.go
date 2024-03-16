package orders_test

import (
	"log"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/common"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/orders"
	"github.com/fr0ster/go-trading-utils/utils"
)

func TestNewLimitOrder(t *testing.T) {
	api_key := os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	secret_key := os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	// Create a new limit order
	order, err := orders.NewLimitOrder(client, "SUSHIUSDT", futures.SideTypeBuy, "5.0", "1.5", futures.TimeInForceTypeGTC)
	if err != nil {
		if apiErr, _ := err.(*common.APIError); apiErr.Code == 0 {
			log.Printf("Error with code 0: %v", err)
			return
		} else {
			log.Fatalf("Error creating limit order: %v", err)
		}
	}

	// Verify the order details
	if order.Symbol != "SUSHIUSDT" {
		t.Errorf("Expected symbol to be SUSHIUSDT, got %s", order.Symbol)
	}
	if order.Side != futures.SideTypeBuy {
		t.Errorf("Expected side to be Buy, got %s", order.Side)
	}
	if order.ExecutedQuantity != "5.0" && order.Status == futures.OrderStatusTypeFilled {
		t.Errorf("Expected quantity to be 5.0, got %s", order.ExecutedQuantity)
	}
	if utils.ConvFloat64ToStr(utils.ConvStrToFloat64(order.Price), 1) != "1.5" {
		t.Errorf("Expected price to be 1.5, got %s", order.Price)
	}
	_, err = orders.CancelOrder(client, order)
	if err != nil {
		log.Fatalf("Error cancelling order: %v", err)
	}
}
