package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/info"
	exchange_interface "github.com/fr0ster/go-trading-utils/interfaces/info"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
)

const degree = 3

func TestGetExchangeInfo(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.NewExchangeInfo()
	info.Init(exchangeInfo, degree, client)

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
	exchangeInfo := exchange_types.NewExchangeInfo()
	info.Init(exchangeInfo, degree, client)

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
	exchangeInfo := exchange_types.NewExchangeInfo()
	info.Init(exchangeInfo, degree, client)

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
	exchangeInfo := exchange_types.NewExchangeInfo()
	info.Init(exchangeInfo, degree, client)

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the permissions is not empty
	if symbol == nil {
		t.Error("GetPermissions returned empty permissions")
	}
}

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.NewExchangeInfo()
	info.Init(exchangeInfo, degree, client)

	test := func(exchangeInfo exchange_interface.ExchangeInfo) {
		_ = exchangeInfo.GetSymbol("BTCUSDT")
		_ = exchangeInfo.GetTimezone()
		_ = exchangeInfo.GetServerTime()
		_ = exchangeInfo.GetRateLimits()
		_ = exchangeInfo.GetExchangeFilters()
	}
	test(exchangeInfo)

}
