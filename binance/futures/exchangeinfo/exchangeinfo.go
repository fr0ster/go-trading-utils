package info

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_types "github.com/fr0ster/go-trading-utils/types/symbols"
	"github.com/fr0ster/go-trading-utils/utils"
)

func InitCreator(client *futures.Client, degree int, symbols ...string) func(val *exchange_types.ExchangeInfo) types.InitFunction {
	return func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
			exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
			if err != nil {
				return
			}
			val.Timezone = exchangeInfo.Timezone
			val.ServerTime = exchangeInfo.ServerTime
			val.RateLimits = ConvertRateLimits(exchangeInfo.RateLimits)
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
			val.Symbols, err = symbols_types.New(
				degree,
				func() (symbols []*symbol_types.Symbol) {
					for _, s := range restrictedSymbols {
						orderTypes := make([]symbol_types.OrderType, len(s.OrderType))
						for i, ot := range s.OrderType {
							orderTypes[i] = symbol_types.OrderType(ot)
						}
						symbols = append(symbols, symbol_types.New(
							s.Symbol,
							items_types.ValueType(utils.ConvStrToFloat64(s.MinNotionalFilter().Notional)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().StepSize)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().MaxQuantity)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().MinQuantity)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().TickSize)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().MaxPrice)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().MinPrice)),
							symbol_types.QuoteAsset(s.QuoteAsset),
							symbol_types.BaseAsset(s.BaseAsset),
							false,
							nil,
							orderTypes,
						))
					}
					return
				})
			return
		}
	}
}

func ConvertRateLimits(rateLimits []futures.RateLimit) []exchange_types.RateLimit {
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
