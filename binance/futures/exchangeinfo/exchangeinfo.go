package info

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
)

func Init(val *exchange_types.ExchangeInfo, degree int, client *futures.Client) error {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return err
	}
	val.Timezone = exchangeInfo.Timezone
	val.ServerTime = exchangeInfo.ServerTime
	val.RateLimits = convertRateLimits(exchangeInfo.RateLimits)
	val.ExchangeFilters = exchangeInfo.ExchangeFilters
	val.Symbols, err = symbols_info.NewSymbols(degree, convertSymbols(exchangeInfo.Symbols))
	return err
}

func RestrictedInit(val *exchange_types.ExchangeInfo, degree int, symbols []string, client *futures.Client) error {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return err
	}
	val.Timezone = exchangeInfo.Timezone
	val.ServerTime = exchangeInfo.ServerTime
	val.RateLimits = convertRateLimits(exchangeInfo.RateLimits)
	val.ExchangeFilters = exchangeInfo.ExchangeFilters
	restrictedSymbols := make([]futures.Symbol, 0)
	symbolMap := make(map[string]bool)
	for _, symbol := range symbols {
		symbolMap[symbol] = true
	}
	for _, s := range exchangeInfo.Symbols {
		if _, exists := symbolMap[s.Symbol]; exists {
			restrictedSymbols = append(restrictedSymbols, s)
		}
	}
	val.Symbols, err = symbols_info.NewSymbols(degree, convertSymbols(restrictedSymbols))
	return err
}

func convertSymbols(symbols []futures.Symbol) []interface{} {
	convertedSymbols := make([]interface{}, len(symbols))
	for i, s := range symbols {
		convertedSymbols[i] = s
	}
	return convertedSymbols
}

func convertRateLimits(rateLimits []futures.RateLimit) []exchange_types.RateLimit {
	convertedRateLimits := make([]exchange_types.RateLimit, len(rateLimits))
	for i, rl := range rateLimits {
		convertedRateLimits[i] = exchange_types.RateLimit{
			RateLimitType: rl.RateLimitType,
			Interval:      rl.Interval,
			Limit:         rl.Limit,
		}
	}
	return convertedRateLimits
}
