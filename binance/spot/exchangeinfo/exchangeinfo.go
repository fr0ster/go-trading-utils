package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_types "github.com/fr0ster/go-trading-utils/types/symbols"
	"github.com/fr0ster/go-trading-utils/utils"
)

type ExchangeInfo exchange_types.ExchangeInfo

func InitCreator(client *binance.Client, degree int, symbol ...string) func(val *exchange_types.ExchangeInfo) types.InitFunction {
	return func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
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
			val.Symbols, err = symbols_types.New(
				degree,
				func() (symbols []*symbol_types.Symbol) {
					for _, s := range exchangeInfo.Symbols {
						orderTypes := make([]symbol_types.OrderType, len(s.OrderTypes))
						for i, ot := range s.OrderTypes {
							orderTypes[i] = symbol_types.OrderType(ot)
						}
						symbols = append(symbols, symbol_types.New(
							s.Symbol,
							items_types.ValueType(utils.ConvStrToFloat64(s.NotionalFilter().MinNotional)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().StepSize)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().MaxQuantity)),
							items_types.QuantityType(utils.ConvStrToFloat64(s.LotSizeFilter().MinQuantity)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().TickSize)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().MaxPrice)),
							items_types.PriceType(utils.ConvStrToFloat64(s.PriceFilter().MinPrice)),
							symbol_types.QuoteAsset(s.QuoteAsset),
							symbol_types.BaseAsset(s.BaseAsset),
							s.IsMarginTradingAllowed,
							s.Permissions,
							orderTypes,
						))
					}
					return
				})
			return err
		}
	}
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
