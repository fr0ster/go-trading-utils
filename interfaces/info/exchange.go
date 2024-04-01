package info

import (
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbols_info "github.com/fr0ster/go-trading-utils/types/info/symbols"
	symbol_info "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

type (
	ExchangeInfo interface {
		GetSymbol(symbol string) *symbol_info.Symbol
		GetSymbols() *symbols_info.Symbols
		GetTimezone() string
		GetServerTime() int64
		GetRateLimits() *exchange_types.RateLimits
		GetExchangeFilters() []interface{}
	}
)
