package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	exchangeinfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	exchange_interface "github.com/fr0ster/go-trading-utils/interfaces/exchangeinfo"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/stretchr/testify/assert"
)

const (
	API_KEY      = "SPOT_TEST_BINANCE_API_KEY"
	SECRET_KEY   = "SPOT_TEST_BINANCE_SECRET_KEY"
	USE_TEST_NET = true
	degree       = 3
)

func TestGetExchangeInfo(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(degree, client))

	// Check if the exchangeInfo is not nil
	if exchangeInfo == nil {
		t.Error("GetExchangeInfo returned nil exchangeInfo")
	}
}

func TestGetOrderTypes(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(degree, client))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: "BTCUSDT"}).(*symbol_info.SpotSymbol)
	orderTypes := symbol.OrderTypes

	// Check if the orderTypes is not empty
	if len(orderTypes) == 0 {
		t.Error("GetOrderTypes returned empty orderTypes")
	}
}

func TestGetPermissions(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(degree, client))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: "BTCUSDT"}).(*symbol_info.SpotSymbol)
	permissions := symbol.Permissions
	assert.NotNil(t, permissions)
}

func TestGetExchangeInfoSymbol(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(degree, client))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: "BTCUSDT"}).(*symbol_info.SpotSymbol)

	// Check if the permissions is not empty
	if symbol == nil {
		t.Error("GetPermissions returned empty permissions")
	}
}

func TestInterface(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(degree, client))

	test := func(exchangeInfo exchange_interface.ExchangeInfo) {
		symbol := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: "BTCUSDT"}).(*symbol_info.SpotSymbol)
		for _, permissions := range symbol.Permissions {
			_ = permissions
		}
		_ = exchangeInfo.GetTimezone()
		_ = exchangeInfo.GetServerTime()
		_ = exchangeInfo.GetRateLimits()
		_ = exchangeInfo.GetExchangeFilters()
	}
	test(exchangeInfo)

}
