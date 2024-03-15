package info

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	symbols_info "github.com/fr0ster/go-binance-utils/futures/info/symbols"
	symbol_info "github.com/fr0ster/go-binance-utils/futures/info/symbols/symbol"
)

type (
	ExchangeInfo struct {
		Timezone        string              `json:"timezone"`
		ServerTime      int64               `json:"serverTime"`
		RateLimits      []futures.RateLimit `json:"rateLimits"`
		ExchangeFilters []interface{}       `json:"exchangeFilters"`
		Symbols         *symbols_info.Symbols
	}
)

func GetExchangeInfo(client *futures.Client) (*ExchangeInfo, error) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	symbols := symbols_info.NewSymbols(2, exchangeInfo.Symbols)
	return &ExchangeInfo{
		Timezone:        exchangeInfo.Timezone,
		ServerTime:      exchangeInfo.ServerTime,
		RateLimits:      exchangeInfo.RateLimits,
		ExchangeFilters: exchangeInfo.ExchangeFilters,
		Symbols:         symbols,
	}, nil
}

func (exchangeInfo *ExchangeInfo) GetSymbol(symbol string) *symbol_info.Symbol {
	return exchangeInfo.Symbols.GetSymbol(symbol)
}
