package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbols_info "github.com/fr0ster/go-trading-utils/types/symbols"
)

type ExchangeInfo exchange_types.ExchangeInfo

func Init(val *exchange_types.ExchangeInfo, degree int, client *binance.Client, symbol ...string) (err error) {
	var (
		exchangeInfo *binance.ExchangeInfo
	)
	if len(symbol) == 0 {
		exchangeInfo, err = client.NewExchangeInfoService().Do(context.Background())
	} else if len(symbol) == 1 {
		exchangeInfo, err = client.NewExchangeInfoService().Symbol(symbol[0]).Do(context.Background())
	} else {
		exchangeInfo, err = client.NewExchangeInfoService().Symbols(symbol...).Do(context.Background())
	}
	if err != nil {
		return err
	}
	val.Timezone = exchangeInfo.Timezone
	val.ServerTime = exchangeInfo.ServerTime
	val.RateLimits = convertRateLimits(exchangeInfo.RateLimits)
	val.ExchangeFilters = exchangeInfo.ExchangeFilters
	val.Symbols, err = symbols_info.NewSymbols(degree, convertSymbols(exchangeInfo.Symbols))
	val.SpotSymbols = exchangeInfo.Symbols
	return err
}

func RestrictedInit(val *exchange_types.ExchangeInfo, degree int, symbols []string, client *binance.Client) error {
	exchangeInfo, err := client.NewExchangeInfoService().Symbols(symbols...).Do(context.Background())
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

func convertSymbols(symbols []binance.Symbol) []interface{} {
	convertedSymbols := make([]interface{}, len(symbols))
	for i, s := range symbols {
		convertedSymbols[i] = s
	}
	return convertedSymbols
}

func convertRateLimits(rateLimits []binance.RateLimit) []exchange_types.RateLimit {
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
