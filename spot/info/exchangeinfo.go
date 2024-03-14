package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

type ExchangeInfo struct {
	exchangeInfo *binance.ExchangeInfo
	Symbols      *Symbols
}

func GetExchangeInfo(client *binance.Client) (ExchangeInfo, error) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return ExchangeInfo{}, err
	}
	symbols := NewSymbols(2)
	for _, symbol := range exchangeInfo.Symbols {
		symbols.Insert(&Symbol{Symbol: symbol.Symbol})
	}
	return ExchangeInfo{exchangeInfo, symbols}, nil
}

func (exchangeInfo *ExchangeInfo) GetOrderTypes(symbolname string) []binance.OrderType {
	res := make([]binance.OrderType, 0)
	for _, info := range exchangeInfo.exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, orderType := range info.OrderTypes {
				res = append(res, binance.OrderType(orderType))
			}
		}
	}
	return res
}

func (exchangeInfo *ExchangeInfo) GetPermissions(symbolname string) []string {
	for _, info := range exchangeInfo.exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			return info.Permissions
		}
	}
	return nil
}

func (exchangeInfo *ExchangeInfo) GetFilters(symbolname string) []map[string]interface{} {
	for _, info := range exchangeInfo.exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			return info.Filters
		}
	}
	return nil
}

func GetFilters_Price_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.PriceFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "PRICE_FILTER" {
					priceFilter := binance.PriceFilter{
						MinPrice: filter["minPrice"].(string),
						MaxPrice: filter["maxPrice"].(string),
						TickSize: filter["tickSize"].(string),
					}
					return &priceFilter
				}
			}
		}
	}
	return nil
}

func GetFilers_Lot_Size_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.LotSizeFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "LOT_SIZE" {
					lotSizeFilter := binance.LotSizeFilter{
						MinQuantity: filter["minQty"].(string),
						MaxQuantity: filter["maxQty"].(string),
						StepSize:    filter["stepSize"].(string),
					}
					return &lotSizeFilter
				}
			}
		}
	}
	return nil
}

func GetFilters_Iceberg_Parts_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.IcebergPartsFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "ICEBERG_PARTS" {
					icebergPartsFilter := binance.IcebergPartsFilter{
						Limit: int(filter["limit"].(float64)),
					}
					return &icebergPartsFilter
				}
			}
		}
	}
	return nil
}

func GetFilters_Market_Lot_Size_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.MarketLotSizeFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "MARKET_LOT_SIZE" {
					marketLotSizeFilter := binance.MarketLotSizeFilter{
						MinQuantity: filter["minQty"].(string),
						MaxQuantity: filter["maxQty"].(string),
						StepSize:    filter["stepSize"].(string),
					}
					return &marketLotSizeFilter
				}
			}
		}
	}
	return nil
}

type TrailingDelta struct {
	minTrailingAboveDelta float64
	minTrailingBelowDelta float64
	maxTrailingAboveDelta float64
	maxTrailingBelowDelta float64
}

func GetFilters_Trailing_Delta_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) TrailingDelta {
	trailingStopFilter := TrailingDelta{}
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "TRAILING_STOP" {
					trailingStopFilter = TrailingDelta{
						minTrailingAboveDelta: filter["minTrailingStop"].(float64),
						minTrailingBelowDelta: filter["minTrailingStop"].(float64),
						maxTrailingAboveDelta: filter["maxTrailingStop"].(float64),
						maxTrailingBelowDelta: filter["maxTrailingStop"].(float64),
					}
					return trailingStopFilter
				}
			}
		}
	}
	return TrailingDelta{}
}

type Percent_Price_By_Side struct {
	bidMultiplierUp   string
	bidMultiplierDown string
	askMultiplierUp   string
	askMultiplierDown string
	averagePriceMins  int
}

func GetFilters_Percent_Price_By_Side_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string, side string) *Percent_Price_By_Side {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "PERCENT_PRICE_BY_SIDE" {
					percentPriceFilter := Percent_Price_By_Side{
						bidMultiplierUp:   filter["bidMultiplierUp"].(string),
						bidMultiplierDown: filter["bidMultiplierDown"].(string),
						askMultiplierUp:   filter["askMultiplierUp"].(string),
						askMultiplierDown: filter["askMultiplierDown"].(string),
						averagePriceMins:  int(filter["averagePriceMins"].(float64)),
					}
					return &percentPriceFilter
				}
			}
		}
	}
	return nil
}

func GetFilters_Notional_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.NotionalFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "NOTIONAL" {
					notionalFilter := binance.NotionalFilter{
						MinNotional: filter["minNotional"].(string),
					}
					return &notionalFilter
				}
			}
		}
	}
	return nil
}

func GetFilters_Max_Num_Orders_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) float64 {
	maxNumOrdersFilter := 0.0
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "MAX_NUM_ORDERS" {
					maxNumOrdersFilter = float64(filter["maxNumOrders"].(float64))
					return maxNumOrdersFilter
				}
			}
		}
	}
	return 0
}

func GetFilters_Max_Num_Algo_Orders_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.MaxNumAlgoOrdersFilter {
	for _, info := range exchangeInfo.Symbols {
		if info.Symbol == symbolname {
			for _, filter := range info.Filters {
				if filter["filterType"] == "MAX_NUM_ALGO_ORDERS" {
					maxNumAlgoOrdersFilter := binance.MaxNumAlgoOrdersFilter{
						MaxNumAlgoOrders: int(filter["maxNumAlgoOrders"].(float64)),
					}
					return &maxNumAlgoOrdersFilter
				}
			}
		}
	}
	return nil
}
