package symbol_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spotInfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	exchange_info "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/stretchr/testify/assert"
)

const degree = 3

func TestNewSymbol(t *testing.T) {
	symbol := &binance.Symbol{
		Symbol: "BTCUSDT",
	}

	s := symbol_info.NewSymbol(symbol)

	if s.Symbol != "BTCUSDT" {
		t.Errorf("Expected SymbolName to be 'BTCUSDT', got %s", s.Symbol)
	}

	// Add more assertions for other fields if needed
}

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)
	exchangeInfo := exchange_info.NewExchangeInfo()
	err := spotInfo.Init(exchangeInfo, degree, client)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	symbol := exchangeInfo.GetSymbol("BTCUSDT")

	// Check if the struct implements the interface
	test := func(s *symbol_info.Symbol) interface{} {
		return s.GetFilter("MAX_NUM_ALGO_ORDERS")
	}
	res := test(symbol)
	assert.NotNil(t, res)
}

// func TestLotSizeFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.LotSizeFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestPriceFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.PriceFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestNotionalFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.NotionalFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestPercentPriceBySideFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.PercentPriceBySideFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestIcebergPartsFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.IcebergPartsFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestMarketLotSizeFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.MarketLotSizeFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestMaxNumOrdersFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.MaxNumOrdersFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestMaxNumAlgoOrdersFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.MaxNumAlgoOrdersFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }

// func TestTrailingDeltaFilter(t *testing.T) {
// 	api_key := os.Getenv("API_KEY")
// 	secret_key := os.Getenv("SECRET_KEY")
// 	// binance.UseTestnet = true
// 	client := binance.NewClient(api_key, secret_key)
// 	exchangeInfo := exchange_info.NewExchangeInfo()
// 	err := spotInfo.Init(exchangeInfo, degree, client)
// 	if err != nil {
// 		t.Errorf("Error: %v", err)
// 	}
// 	symbol := exchangeInfo.GetSymbol("BTCUSDT")
// 	lotSizeFilter := symbol.TrailingDeltaFilter()
// 	assert.NotNil(t, lotSizeFilter)
// }
