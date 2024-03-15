package symbol_test

import (
	"testing"

	"github.com/adshao/go-binance/v2"
	symbol_info "github.com/fr0ster/go-binance-utils/spot/info/symbols/symbol"
)

func TestNewSymbol(t *testing.T) {
	symbol := &binance.Symbol{
		Symbol: "BTCUSDT",
	}

	s := symbol_info.NewSymbol(2, symbol)

	if s.SymbolName != "BTCUSDT" {
		t.Errorf("Expected SymbolName to be 'BTCUSDT', got %s", s.SymbolName)
	}

	// Add more assertions for other fields if needed
}
