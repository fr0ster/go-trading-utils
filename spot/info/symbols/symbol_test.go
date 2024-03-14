package symbols_test

import (
	"testing"

	symbol_info "github.com/fr0ster/go-binance-utils/spot/info/symbols"
)

func TestSymbols_Insert(t *testing.T) {
	symbols := symbol_info.NewSymbols(3)

	symbol := &symbol_info.Symbol{
		Symbol: "BTCUSDT",
	}

	symbols.Insert(symbol)

	if symbols.Len() != 1 {
		t.Errorf("Expected symbols length to be 1, got %d", symbols.Len())
	}

	if insertedSymbol := symbols.GetSymbol("BTCUSDT"); insertedSymbol != symbol {
		t.Errorf("Expected inserted symbol to be %v, got %v", symbol, insertedSymbol)
	}
}

func TestSymbols_DeleteSymbol(t *testing.T) {
	symbols := symbol_info.NewSymbols(3)

	symbol := &symbol_info.Symbol{
		Symbol: "BTCUSDT",
	}

	symbols.Insert(symbol)

	symbols.DeleteSymbol("BTCUSDT")

	if symbols.Len() != 0 {
		t.Errorf("Expected symbols length to be 0, got %d", symbols.Len())
	}

	if deletedSymbol := symbols.GetSymbol("BTCUSDT"); deletedSymbol != nil {
		t.Errorf("Expected deleted symbol to be nil, got %v", deletedSymbol)
	}
}

// Add more tests for other functions...
