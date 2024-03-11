package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
)

func TestGetBookTicker(t *testing.T) {
	// Create a mock book ticker
	bookTicker := binance.BookTicker{
		Symbol:      "BTCUSDT",
		BidPrice:    "10000",
		BidQuantity: "1",
		AskPrice:    "10001",
		AskQuantity: "1",
	}

	// Set the mock book ticker in the book ticker map
	info.SetBookTickerMapItem("BTCUSDT", bookTicker)

	// Get the book ticker from the map
	result := info.GetBookTickerMapItem("BTCUSDT")

	// Check if the retrieved book ticker matches the expected value
	if result != bookTicker {
		t.Errorf("Expected book ticker: %v, got: %v", bookTicker, result)
	}
}

func TestInitPricesMap(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the InitPricesMap function
	err := info.InitPricesMap(client, "BTCUSDT")

	// Check if there was an error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Get the book ticker map
	bookTickerMap := info.GetBookTickerMap()

	// Check if the book ticker map is not empty
	if len(bookTickerMap) == 0 {
		t.Errorf("Expected non-empty book ticker map, got empty map")
	}
}

// Add more tests for other functions if needed
