package exchangeinfo

import (
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
)

type (
	RateLimits struct {
		Interval    time.Duration
		IntervalNum int64
		Limit       int64
	}
	RateLimit struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int64  `json:"intervalNum"`
		Limit         int64  `json:"limit"`
	}
	ExchangeInfo struct {
		Timezone        string                `json:"timezone"`
		ServerTime      int64                 `json:"serverTime"`
		RateLimits      []RateLimit           `json:"rateLimits"`
		ExchangeFilters []interface{}         `json:"exchangeFilters"`
		Symbols         *symbols_info.Symbols `json:"symbols"`
	}
)

// GetExchangeFilters implements info.ExchangeInfo.
func (e *ExchangeInfo) GetExchangeFilters() []interface{} {
	return e.ExchangeFilters
}

func (e *ExchangeInfo) GetRateLimits() *[]RateLimit {
	res := append([]RateLimit{}, e.RateLimits...)
	return &res
}

func (e *ExchangeInfo) get_limit(rateLimitType, interval string) *RateLimits {
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == rateLimitType || rateLimit.Interval == interval {
			return &RateLimits{
				Interval:    time.Minute,
				IntervalNum: rateLimit.IntervalNum,
				Limit:       rateLimit.Limit,
			}
		}
	}
	return nil
}

func (e *ExchangeInfo) Get_Minute_Request_Limit() *RateLimits {
	return e.get_limit("REQUEST_WEIGHT", "MINUTE")
}

func (e *ExchangeInfo) Get_Minute_Order_Limit() *RateLimits {
	return e.get_limit("ORDERS", "MINUTE")
}

func (e *ExchangeInfo) Get_Day_Order_Limit() *RateLimits {
	return e.get_limit("ORDERS", "DAY")
}

func (e *ExchangeInfo) Get_Minute_Raw_Request_Limit() *RateLimits {
	return e.get_limit("RAW_REQUESTS", "MINUTE")
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

func New(init func(*ExchangeInfo) types.InitFunction) *ExchangeInfo {
	this := &ExchangeInfo{
		Timezone:   "",
		ServerTime: 0,
		RateLimits: nil,
		Symbols:    nil,
	}
	if init != nil {
		init(this)()
	}
	return this
}
