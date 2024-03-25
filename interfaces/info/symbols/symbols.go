package symbols

import (
	symbol_info "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

type (
	Symbols interface {
		Lock()
		Unlock()
		GetSymbol(symbol string) symbol_info.Symbol
		Insert(symbol symbol_info.Symbol)
		Len() int
	}
)
