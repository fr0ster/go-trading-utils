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
		GetRateLimits() *[]exchange_types.RateLimit
		Get_Minute_Request_Limit() *exchange_types.RateLimits
		Get_Minute_Order_Limit() *exchange_types.RateLimits
		Get_Day_Order_Limit() *exchange_types.RateLimits
		Get_Minute_Raw_Request_Limit() *exchange_types.RateLimits
		GetExchangeFilters() []interface{}
	}
)
