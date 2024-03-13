package services_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	"github.com/fr0ster/go-binance-utils/spot/services"
)

func TestGetMarketPrice(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Define test cases
	testCases := []struct {
		symbol        string
		expectedPrice float64
		expectedErr   error
	}{
		{
			symbol:        "BTCUSDT",
			expectedPrice: 50000.0,
			expectedErr:   nil,
		},
		{
			symbol:        "ETHUSDT",
			expectedPrice: 2000.0,
			expectedErr:   nil,
		},
		{
			symbol:        "INVALID",
			expectedPrice: 0.0,
			expectedErr:   common.APIError{Code: -1121, Message: "Invalid symbol."},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		_, _, err := services.GetMarketPrice(client, tc.symbol)

		if err != nil && tc.expectedErr == nil {
			t.Errorf("Expected no error, but got %v", err)
		} else if err != nil && err.Error() != tc.expectedErr.Error() {
			t.Errorf("Expected error %v, but got %v", tc.expectedErr, err)
		}
	}
}
