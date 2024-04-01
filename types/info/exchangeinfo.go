package info

import (
	"time"

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

func (e *ExchangeInfo) GetRateLimits() *[]RateLimit {
	res := make([]RateLimit, 0)
	for _, rateLimit := range e.RateLimits {
		res = append(res, rateLimit)
	}
	return &res
}

func (e *ExchangeInfo) Get_Minute_Request_Limit() (time.Duration, int64, int64) {
	var timeDuration time.Duration
	var intervalNum int64
	var limit int64
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "REQUEST_WEIGHT" || rateLimit.Interval == "MINUTE" {
			timeDuration = time.Minute
			intervalNum = rateLimit.IntervalNum
			limit = rateLimit.Limit
			break
		}
	}
	return timeDuration, intervalNum, limit
}

func (e *ExchangeInfo) Get_Minute_Order_Limit() (time.Duration, int64, int64) {
	var timeDuration time.Duration
	var intervalNum int64
	var limit int64
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "ORDERS" || rateLimit.Interval == "MINUTE" {
			timeDuration = time.Minute
			intervalNum = rateLimit.IntervalNum
			limit = rateLimit.Limit
			break
		}
	}
	return timeDuration, intervalNum, limit
}

func (e *ExchangeInfo) Get_Day_Order_Limit() (time.Duration, int64, int64) {
	var timeDuration time.Duration
	var intervalNum int64
	var limit int64
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "ORDERS" || rateLimit.Interval == "DAY" {
			timeDuration = time.Hour * 24
			intervalNum = rateLimit.IntervalNum
			limit = rateLimit.Limit
			break
		}
	}
	return timeDuration, intervalNum, limit
}

func (e *ExchangeInfo) Get_Minute_Raw_Request_Limit() (time.Duration, int64, int64) {
	var timeDuration time.Duration
	var intervalNum int64
	var limit int64
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "RAW_REQUESTS" || rateLimit.Interval == "MINUTE" {
			timeDuration = time.Minute
			intervalNum = rateLimit.IntervalNum
			limit = rateLimit.Limit
			break
		}
	}
	return timeDuration, intervalNum, limit
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
