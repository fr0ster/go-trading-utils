package symbol_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/info"
	symbol_info "github.com/fr0ster/go-trading-utils/binance/futures/info/symbols/symbol"
	"github.com/stretchr/testify/assert"
)

func TestNewSymbol(t *testing.T) {
	symbol := &futures.Symbol{
		Symbol: "BTCUSDT",
	}

	s := symbol_info.NewSymbol(2, symbol)

	if s.SymbolName != "BTCUSDT" {
		t.Errorf("Expected SymbolName to be 'BTCUSDT', got %s", s.SymbolName)
	}

	// Add more assertions for other fields if needed
}

func TestInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)
	exchangeInfo, err := exchange_info.NewExchangeInfo(client)
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
