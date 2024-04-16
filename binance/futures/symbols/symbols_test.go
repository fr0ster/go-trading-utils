package symbols_test

import (
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
	"github.com/stretchr/testify/assert"
)

func TestSymbolsNew(t *testing.T) {
	symbols, err := symbols_info.NewSymbols(2, []interface{}{})
	assert.Nil(t, err)

	if symbols == nil {
		t.Errorf("Expected symbols to be initialized, but got nil")
	}
}

func TestSymbolsLen(t *testing.T) {
	symbols, err :=
		symbols_info.NewSymbols(2, []interface{}{futures.Symbol{Symbol: "BTCUSDT"}})
	assert.Nil(t, err)

	// TODO: Add test cases to insert symbols into the BTree

	expectedLen := 1 // Replace with the expected length
	actualLen := symbols.Len()

	if actualLen != expectedLen {
		t.Errorf("Expected Len to be %d, but got %d", expectedLen, actualLen)
	}
}

func TestSymbolsInsert(t *testing.T) {
	symbols, err :=
		symbols_info.NewSymbols(2, []interface{}{})
	assert.Nil(t, err)

	symbol := &symbol_info.FuturesSymbol{
		Symbol: "BTCUSDT",
	}

	symbols.Insert(symbol)

	// TODO: Add assertions or additional tests
}

func TestSymbolsGetSymbol(t *testing.T) {
	symbolName := "BTCUSDT"
	symbols, err :=
		symbols_info.NewSymbols(2, append([]interface{}{}, func() []interface{} {
			var symbols []interface{}
			for _, symbol := range []futures.Symbol{{Symbol: symbolName}} {
				symbols = append(symbols, symbol)
			}
			return symbols
		}()...))
	assert.Nil(t, err)

	// TODO: Add test cases to insert symbols into the BTree

	actualSymbol := symbols.GetSymbol(&symbol_info.FuturesSymbol{Symbol: symbolName})

	if actualSymbol == nil {
		t.Errorf("Expected to get symbol %s, but got nil", symbolName)
	}

	// TODO: Add assertions to compare the actual symbol with the expected symbol
}
