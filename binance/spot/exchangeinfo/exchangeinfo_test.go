package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	exchangeinfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
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

	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(client, degree, "BTCUSDT"))

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
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(client, degree, "BTCUSDT"))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the orderTypes is not empty
	assert.NotNil(t, symbol)
}

func TestGetPermissions(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(client, degree, "BTCUSDT"))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")
	permissions := symbol.GetPermissions()
	assert.NotNil(t, permissions)
}

func TestGetExchangeInfoSymbol(t *testing.T) {
	api_key := os.Getenv(API_KEY)
	secret_key := os.Getenv(SECRET_KEY)
	binance.UseTestnet = USE_TEST_NET
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(exchangeinfo.InitCreator(client, degree, "BTCUSDT"))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the permissions is not empty
	if symbol == nil {
		t.Error("GetPermissions returned empty permissions")
	}
}
