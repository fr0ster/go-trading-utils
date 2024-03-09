package services

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
)

func TestGetExchangeInfo(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the function being tested
	exchangeInfo, err := GetExchangeInfo(client)

	// Check if the function returned an error
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Check if the exchangeInfo is not nil
	if exchangeInfo == nil {
		t.Error("GetExchangeInfo returned nil exchangeInfo")
	}
}

func TestGetOrderTypes(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo, err := GetExchangeInfo(client)
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Call the function being tested
	orderTypes := GetOrderTypes(exchangeInfo, "BTCUSDT")

	// Check if the orderTypes is not empty
	if len(orderTypes) == 0 {
		t.Error("GetOrderTypes returned empty orderTypes")
	}
}

func TestGetPermissions(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo, err := GetExchangeInfo(client)
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Call the function being tested
	permissions := GetPermissions(exchangeInfo, "BTCUSDT")

	// Check if the permissions is not empty
	if len(permissions) == 0 {
		t.Error("GetPermissions returned empty permissions")
	}
}
