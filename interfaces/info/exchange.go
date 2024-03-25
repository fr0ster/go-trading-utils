package info

import (
	symbols_interface "github.com/fr0ster/go-trading-utils/interfaces/info/symbols"
)

type (
	ExchangeInfo interface {
		GetSymbol(symbol string) symbols_interface.Symbols
	}
)
