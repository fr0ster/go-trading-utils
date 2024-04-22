package exchangeinfo

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
	"github.com/google/btree"
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
		SpotSymbols     []binance.Symbol      `json:"spotSymbols"`
		FuturesSymbols  []futures.Symbol      `json:"futuresSymbols"`
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

func (exchangeInfo *ExchangeInfo) GetSymbol(symbol btree.Item) btree.Item {
	return exchangeInfo.Symbols.GetSymbol(symbol)
}

func (exchangeInfo *ExchangeInfo) GetSymbols() *symbols_info.Symbols {
	return exchangeInfo.Symbols
}

// Ascend implements info.ExchangeInfo.
func (exchangeInfo *ExchangeInfo) Ascend(iterator func(item btree.Item) bool) {
	exchangeInfo.Symbols.Ascend(iterator)
}

// Descend implements info.ExchangeInfo.
func (exchangeInfo *ExchangeInfo) Descend(iterator func(item btree.Item) bool) {
	exchangeInfo.Symbols.Descend(iterator)
}

func New() *ExchangeInfo {
	return &ExchangeInfo{
		Timezone:   "",
		ServerTime: 0,
		RateLimits: nil,
		Symbols:    nil,
	}
}
