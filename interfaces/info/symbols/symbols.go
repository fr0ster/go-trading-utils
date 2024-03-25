package symbols

import (
	symbol_interface "github.com/fr0ster/go-trading-utils/interfaces/info/symbols/symbol"
)

type (
	Symbols interface {
		Lock()
		Unlock()
		GetSymbol(symbol string) symbol_interface.Symbol
		Insert(symbol symbol_interface.Symbol)
		Len() int
	}
)
