package symbols_test

import (
	"testing"

	"github.com/adshao/go-binance/v2"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
)

func TestSymbolsNew(t *testing.T) {
	symbols := symbols_info.NewSymbols(2, []interface{}{})

	if symbols == nil {
		t.Errorf("Expected symbols to be initialized, but got nil")
	}
}

func TestSymbolsLen(t *testing.T) {
	symbols :=
		symbols_info.NewSymbols(2, []interface{}{binance.Symbol{Symbol: "BTCUSDT"}})

	// TODO: Add test cases to insert symbols into the BTree

	expectedLen := 1 // Replace with the expected length
	actualLen := symbols.Len()

	if actualLen != expectedLen {
		t.Errorf("Expected Len to be %d, but got %d", expectedLen, actualLen)
	}
}

func TestSymbolsInsert(t *testing.T) {
	symbols :=
		symbols_info.NewSymbols(2, []interface{}{})

	symbol := symbol_info.NewSymbol(&binance.Symbol{
		// Initialize the symbol with test data
	})

	symbols.Insert(symbol)

	// TODO: Add assertions or additional tests
}

func TestSymbolsGetSymbol(t *testing.T) {
	symbolName := "BTCUSDT"
	symbols :=
		symbols_info.NewSymbols(2, append([]interface{}{}, []interface{}{binance.Symbol{Symbol: symbolName}}...))

	// TODO: Add test cases to insert symbols into the BTree

	actualSymbol := symbols.GetSymbol(symbolName)

	if actualSymbol == nil {
		t.Errorf("Expected to get symbol %s, but got nil", symbolName)
	}

	// TODO: Add assertions to compare the actual symbol with the expected symbol
}
