package orders_test

import (
	"log"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/orders"
	"github.com/fr0ster/go-binance-utils/utils"
)

func TestNewLimitOrder(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	// Create a new limit order
	order, err := orders.NewLimitOrder(client, "SUSHIUSDT", binance.SideTypeBuy, "5.0", "1.5", binance.TimeInForceTypeGTC)
	if err != nil {
		log.Fatalf("Error creating limit order: %v", err)
	}

	// Verify the order details
	if order.Symbol != "SUSHIUSDT" {
		t.Errorf("Expected symbol to be SUSHIUSDT, got %s", order.Symbol)
	}
	if order.Side != binance.SideTypeBuy {
		t.Errorf("Expected side to be Buy, got %s", order.Side)
	}
	if order.ExecutedQuantity != "5.0" && order.Status == binance.OrderStatusTypeFilled {
		t.Errorf("Expected quantity to be 5.0, got %s", order.ExecutedQuantity)
	}
	if utils.ConvFloat64ToStr(utils.ConvStrToFloat64(order.Price), 1) != "1.5" {
		t.Errorf("Expected price to be 1.5, got %s", order.Price)
	}
}
