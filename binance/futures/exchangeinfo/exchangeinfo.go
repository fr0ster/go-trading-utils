package info

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
)

func InitCreator(degree int, client *futures.Client) func(*exchange_types.ExchangeInfo) types.InitFunction {
	return func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
			var (
				exchangeInfo *futures.ExchangeInfo
			)
			exchangeInfo, err = client.NewExchangeInfoService().Do(context.Background())
			if err != nil {
				return
			}
			val.Timezone = exchangeInfo.Timezone
			val.ServerTime = exchangeInfo.ServerTime
			val.RateLimits = convertRateLimits(exchangeInfo.RateLimits)
			val.ExchangeFilters = exchangeInfo.ExchangeFilters
			val.Symbols, err = symbols_info.NewSymbols(degree, convertSymbols(exchangeInfo.Symbols))
			return
		}
	}
}

func RestrictedInitCreator(degree int, symbols []string, client *futures.Client) func(val *exchange_types.ExchangeInfo) types.InitFunction {
	return func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
			exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
			if err != nil {
				return
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
			return
		}
	}
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
