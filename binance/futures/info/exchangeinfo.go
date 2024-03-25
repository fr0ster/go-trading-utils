package info

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	exchange_types "github.com/fr0ster/go-trading-utils/types/info"
	symbols_info "github.com/fr0ster/go-trading-utils/types/info/symbols"
)

func Init(val *exchange_types.ExchangeInfo, client *futures.Client) error {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return err
	}
	val.Timezone = exchangeInfo.Timezone
	val.ServerTime = exchangeInfo.ServerTime
	val.RateLimits = convertRateLimits(exchangeInfo.RateLimits)
	val.ExchangeFilters = exchangeInfo.ExchangeFilters
	val.Symbols = symbols_info.NewSymbols(2, convertSymbols(exchangeInfo.Symbols))
	return nil
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
