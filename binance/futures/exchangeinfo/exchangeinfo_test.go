package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futuresInfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	"github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	"github.com/sirupsen/logrus"
)

const degree = 3

func TestGetExchangeInfo(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)

	exchangeInfo := exchangeinfo.New(futuresInfo.InitCreator(client, degree, "BTCUSDT"))

	// Check if the exchangeInfo is not nil
	if exchangeInfo == nil {
		t.Error("GetExchangeInfo returned nil exchangeInfo")
	}
}

func TestGetOrderTypes(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	exchangeInfo := exchangeinfo.New(futuresInfo.InitCreator(client, degree, "BTCUSDT"))

	// Call the function being tested
	orderTypes := exchangeInfo.GetSymbol("BTCUSDT")
	logrus.Info(orderTypes)
}

func TestGetExchangeInfoSymbol(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	exchangeInfo := exchangeinfo.New(futuresInfo.InitCreator(client, degree, "BTCUSDT"))

	// Call the function being tested
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the permissions is not empty
	if symbol == nil {
		t.Error("GetPermissions returned empty permissions")
	}
}
