package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
)

func TestGetExchangeInfo(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo, err := info.GetExchangeInfo(client)
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
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo, err := info.GetExchangeInfo(client)
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Call the function being tested
	orderTypes := exchangeInfo.GetSymbol("BTCUSDT").OrderTypes

	// Check if the orderTypes is not empty
	if len(orderTypes) == 0 {
		t.Error("GetOrderTypes returned empty orderTypes")
	}
}

func TestGetPermissions(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo, err := info.GetExchangeInfo(client)
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Call the function being tested
	permissions := exchangeInfo.GetSymbol("BTCUSDT").Permissions

	// Check if the permissions is not empty
	if len(permissions) == 0 {
		t.Error("GetPermissions returned empty permissions")
	}
}

func TestGetExchangeInfoSymbol(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo, err := info.GetExchangeInfo(client)
	if err != nil {
		t.Errorf("GetExchangeInfo returned an error: %v", err)
	}

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the permissions is not empty
	if symbol == nil {
		t.Error("GetPermissions returned empty permissions")
	}
}
