package info_test

import (
	"testing"

	"github.com/fr0ster/go-binance-utils/info"
)

func TestSymbols_Insert(t *testing.T) {
	symbols := info.NewSymbols(3)

	symbol := &info.Symbol{
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
	symbols := info.NewSymbols(3)

	symbol := &info.Symbol{
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
