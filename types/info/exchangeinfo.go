package info

import (
	symbols_info "github.com/fr0ster/go-trading-utils/types/info/symbols"
	symbol_info "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
)

type (
	RateLimits struct {
		RequestWeightMinuteIntervalNum int64
		RequestWeightMinuteLimit       int64
		OrdersMinuteIntervalNum        int64
		OrdersMinuteLimit              int64
		OrdersDayIntervalNum           int64
		OrdersDayLimit                 int64
		RawRequestsMinuteNum           int64
		RawRequestsMinuteLimit         int64
	}
	RateLimit struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int64  `json:"intervalNum"`
		Limit         int64  `json:"limit"`
	}
	ExchangeInfo struct {
		Timezone        string        `json:"timezone"`
		ServerTime      int64         `json:"serverTime"`
		RateLimits      []RateLimit   `json:"rateLimits"`
		ExchangeFilters []interface{} `json:"exchangeFilters"`
		Symbols         *symbols_info.Symbols
	}
)

// GetExchangeFilters implements info.ExchangeInfo.
func (e *ExchangeInfo) GetExchangeFilters() []interface{} {
	return e.ExchangeFilters
}

// // GetRateLimits implements info.ExchangeInfo.
// func (e *ExchangeInfo) GetRateLimits() []RateLimit {
// 	return e.RateLimits
// }

func (e *ExchangeInfo) GetRateLimits() *RateLimits {
	res := &RateLimits{}
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "REQUEST_WEIGHT" || rateLimit.Interval == "MINUTE" {
			res.RequestWeightMinuteIntervalNum = rateLimit.IntervalNum
			res.RequestWeightMinuteLimit = rateLimit.Limit
		}
		if rateLimit.RateLimitType == "ORDERS" || rateLimit.Interval == "MINUTE" {
			res.OrdersMinuteIntervalNum = rateLimit.IntervalNum
			res.OrdersMinuteLimit = rateLimit.Limit
		}
		if rateLimit.RateLimitType == "ORDERS" || rateLimit.Interval == "DAY" {
			res.OrdersDayIntervalNum = rateLimit.IntervalNum
			res.OrdersDayLimit = rateLimit.Limit
		}
		if rateLimit.RateLimitType == "RAW_REQUESTS" || rateLimit.Interval == "MINUTE" {
			res.RawRequestsMinuteNum = rateLimit.IntervalNum
			res.RawRequestsMinuteLimit = rateLimit.Limit
		}
	}
	return res
}

// GetServerTime implements info.ExchangeInfo.
func (e *ExchangeInfo) GetServerTime() int64 {
	return e.ServerTime
}

// GetTimezone implements info.ExchangeInfo.
func (e *ExchangeInfo) GetTimezone() string {
	return e.Timezone
}

func (exchangeInfo *ExchangeInfo) GetSymbol(symbol string) *symbol_info.Symbol {
	return exchangeInfo.Symbols.GetSymbol(symbol)
}

func (exchangeInfo *ExchangeInfo) GetSymbols() *symbols_info.Symbols {
	return exchangeInfo.Symbols
}

func NewExchangeInfo() *ExchangeInfo {
	return &ExchangeInfo{
		Timezone:   "",
		ServerTime: 0,
		RateLimits: nil,
		Symbols:    nil,
	}
}
