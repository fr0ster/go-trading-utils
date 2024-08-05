package exchange

import (
	"fmt"
)

const (
	SymbolFilterTypeLotSize            SymbolFilterType = "LOT_SIZE"
	SymbolFilterTypePriceFilter        SymbolFilterType = "PRICE_FILTER"
	SymbolFilterTypePercentPriceBySide SymbolFilterType = "PERCENT_PRICE_BY_SIDE"
	SymbolFilterTypeMinNotional        SymbolFilterType = "MIN_NOTIONAL"
	SymbolFilterTypeNotional           SymbolFilterType = "NOTIONAL"
	SymbolFilterTypeIcebergParts       SymbolFilterType = "ICEBERG_PARTS"
	SymbolFilterTypeMarketLotSize      SymbolFilterType = "MARKET_LOT_SIZE"
	SymbolFilterTypeMaxNumOrders       SymbolFilterType = "MAX_NUM_ORDERS"
	SymbolFilterTypeMaxNumAlgoOrders   SymbolFilterType = "MAX_NUM_ALGO_ORDERS"
	SymbolFilterTypeTrailingDelta      SymbolFilterType = "TRAILING_DELTA"
)

type (
	SymbolFilterType string

	// RateLimit struct
	RateLimit struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int64  `json:"intervalNum"`
		Limit         int64  `json:"limit"`
	}

	// Symbol market symbol
	Symbol struct {
		Symbol  string                   `json:"symbol"`
		Pair    string                   `json:"pair"`
		Filters []map[string]interface{} `json:"filters"`
	}

	// LotSizeFilter define lot size filter of symbol
	LotSizeFilter struct {
		MaxQuantity string `json:"maxQty"`
		MinQuantity string `json:"minQty"`
		StepSize    string `json:"stepSize"`
	}

	// PriceFilter define price filter of symbol
	PriceFilter struct {
		MaxPrice string `json:"maxPrice"`
		MinPrice string `json:"minPrice"`
		TickSize string `json:"tickSize"`
	}

	// PERCENT_PRICE_BY_SIDE define percent price filter of symbol by side
	PercentPriceBySideFilter struct {
		AveragePriceMins  int    `json:"avgPriceMins"`
		BidMultiplierUp   string `json:"bidMultiplierUp"`
		BidMultiplierDown string `json:"bidMultiplierDown"`
		AskMultiplierUp   string `json:"askMultiplierUp"`
		AskMultiplierDown string `json:"askMultiplierDown"`
	}

	// NotionalFilter define notional filter of symbol
	NotionalFilter struct {
		MinNotional      string `json:"minNotional"`
		ApplyMinToMarket bool   `json:"applyMinToMarket"`
		MaxNotional      string `json:"maxNotional"`
		ApplyMaxToMarket bool   `json:"applyMaxToMarket"`
		AvgPriceMins     int    `json:"avgPriceMins"`
	}

	// IcebergPartsFilter define iceberg part filter of symbol
	IcebergPartsFilter struct {
		Limit int `json:"limit"`
	}

	// MarketLotSizeFilter define market lot size filter of symbol
	MarketLotSizeFilter struct {
		MaxQuantity string `json:"maxQty"`
		MinQuantity string `json:"minQty"`
		StepSize    string `json:"stepSize"`
	}

	// Spot trading supports tracking stop orders
	// Tracking stop loss sets an automatic trigger price based on market price using a new parameter trailingDelta
	TrailingDeltaFilter struct {
		MinTrailingAboveDelta int `json:"minTrailingAboveDelta"`
		MaxTrailingAboveDelta int `json:"maxTrailingAboveDelta"`
		MinTrailingBelowDelta int `json:"minTrailingBelowDelta"`
		MaxTrailingBelowDelta int `json:"maxTrailingBelowDelta"`
	}

	// The "Algo" order is STOP_ LOSS, STOP_ LOS_ LIMITED, TAKE_ PROFIT and TAKE_ PROFIT_ Limit Stop Loss Order.
	// Therefore, orders other than the above types are non conditional(Algo) orders, and MaxNumOrders defines the maximum
	// number of orders placed for these types of orders
	MaxNumOrdersFilter struct {
		MaxNumOrders int `json:"maxNumOrders"`
	}

	// MaxNumAlgoOrdersFilter define max num algo orders filter of symbol
	MaxNumAlgoOrdersFilter struct {
		MaxNumAlgoOrders int `json:"maxNumAlgoOrders"`
	}
)

// LotSizeFilter return lot size filter of symbol
func (s *Symbol) LotSizeFilter() *LotSizeFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeLotSize) {
			f := &LotSizeFilter{}
			if i, ok := filter["maxQty"]; ok {
				f.MaxQuantity = i.(string)
			}
			if i, ok := filter["minQty"]; ok {
				f.MinQuantity = i.(string)
			}
			if i, ok := filter["stepSize"]; ok {
				f.StepSize = i.(string)
			}
			return f
		}
	}
	return nil
}

// PriceFilter return price filter of symbol
func (s *Symbol) PriceFilter() *PriceFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypePriceFilter) {
			f := &PriceFilter{}
			if i, ok := filter["maxPrice"]; ok {
				f.MaxPrice = i.(string)
			}
			if i, ok := filter["minPrice"]; ok {
				f.MinPrice = i.(string)
			}
			if i, ok := filter["tickSize"]; ok {
				f.TickSize = i.(string)
			}
			return f
		}
	}
	return nil
}

// PercentPriceBySideFilter return percent price filter of symbol
func (s *Symbol) PercentPriceBySideFilter() *PercentPriceBySideFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypePercentPriceBySide) {
			f := &PercentPriceBySideFilter{}
			if i, ok := filter["avgPriceMins"]; ok {
				if apm, okk := ToInt(i); okk == nil {
					f.AveragePriceMins = apm
				}
			}
			if i, ok := filter["bidMultiplierUp"]; ok {
				f.BidMultiplierUp = i.(string)
			}
			if i, ok := filter["bidMultiplierDown"]; ok {
				f.BidMultiplierDown = i.(string)
			}
			if i, ok := filter["askMultiplierUp"]; ok {
				f.AskMultiplierUp = i.(string)
			}
			if i, ok := filter["askMultiplierDown"]; ok {
				f.AskMultiplierDown = i.(string)
			}
			return f
		}
	}
	return nil
}

// NotionalFilter return notional filter of symbol
func (s *Symbol) NotionalFilter() *NotionalFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeNotional) {
			f := &NotionalFilter{}
			if i, ok := filter["minNotional"]; ok {
				f.MinNotional = i.(string)
			}
			if i, ok := filter["applyMinToMarket"]; ok {
				f.ApplyMinToMarket = i.(bool)
			}
			if i, ok := filter["maxNotional"]; ok {
				f.MaxNotional = i.(string)
			}
			if i, ok := filter["applyMaxToMarket"]; ok {
				f.ApplyMaxToMarket = i.(bool)
			}
			if i, ok := filter["avgPriceMins"]; ok {
				if apm, okk := ToInt(i); okk == nil {
					f.AvgPriceMins = apm
				}
			}
			return f
		}
	}
	return nil
}

// IcebergPartsFilter return iceberg part filter of symbol
func (s *Symbol) IcebergPartsFilter() *IcebergPartsFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeIcebergParts) {
			f := &IcebergPartsFilter{}
			if i, ok := filter["limit"]; ok {
				if limit, okk := ToInt(i); okk == nil {
					f.Limit = limit
				}
			}
			return f
		}
	}
	return nil
}

// MarketLotSizeFilter return market lot size filter of symbol
func (s *Symbol) MarketLotSizeFilter() *MarketLotSizeFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMarketLotSize) {
			f := &MarketLotSizeFilter{}
			if i, ok := filter["maxQty"]; ok {
				f.MaxQuantity = i.(string)
			}
			if i, ok := filter["minQty"]; ok {
				f.MinQuantity = i.(string)
			}
			if i, ok := filter["stepSize"]; ok {
				f.StepSize = i.(string)
			}
			return f
		}
	}
	return nil
}

// For specific meanings, please refer to the type definition MaxNumOrders
func (s *Symbol) MaxNumOrdersFilter() *MaxNumOrdersFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMaxNumOrders) {
			f := &MaxNumOrdersFilter{}
			if i, ok := filter["maxNumOrders"]; ok {
				if mno, okk := ToInt(i); okk == nil {
					f.MaxNumOrders = mno
				}
			}
			return f
		}
	}
	return nil
}

// MaxNumAlgoOrdersFilter return max num algo orders filter of symbol
func (s *Symbol) MaxNumAlgoOrdersFilter() *MaxNumAlgoOrdersFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMaxNumAlgoOrders) {
			f := &MaxNumAlgoOrdersFilter{}
			if i, ok := filter["maxNumAlgoOrders"]; ok {
				if mnao, okk := ToInt(i); okk == nil {
					f.MaxNumAlgoOrders = mnao
				}
			}
			return f
		}
	}
	return nil
}

// For specific meanings, please refer to the type definition TrailingDeltaFilter
func (s *Symbol) TrailingDeltaFilter() *TrailingDeltaFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeTrailingDelta) {
			f := &TrailingDeltaFilter{}
			if i, ok := filter["minTrailingAboveDelta"]; ok {
				if mtad, okk := ToInt(i); okk == nil {
					f.MinTrailingAboveDelta = mtad
				}
			}
			if i, ok := filter["maxTrailingAboveDelta"]; ok {
				if mtad, okk := ToInt(i); okk == nil {
					f.MaxTrailingAboveDelta = mtad
				}
			}
			if i, ok := filter["minTrailingBelowDelta"]; ok {
				if mtbd, okk := ToInt(i); okk == nil {
					f.MinTrailingBelowDelta = mtbd
				}
			}
			if i, ok := filter["maxTrailingBelowDelta"]; ok {
				if mtbd, okk := ToInt(i); okk == nil {
					f.MaxTrailingBelowDelta = mtbd
				}
			}
			return f
		}
	}
	return nil
}

func ToInt(digit interface{}) (i int, err error) {
	if intVal, ok := digit.(int); ok {
		return int(intVal), nil
	}
	if floatVal, ok := digit.(float64); ok {
		return int(floatVal), nil
	}
	return 0, fmt.Errorf("unexpected digit: %v", digit)
}
