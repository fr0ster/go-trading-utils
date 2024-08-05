package exchange

import (
	"fmt"
)

const (
	SymbolFilterTypeLotSize          SymbolFilterType = "LOT_SIZE"
	SymbolFilterTypePrice            SymbolFilterType = "PRICE_FILTER"
	SymbolFilterTypePercentPrice     SymbolFilterType = "PERCENT_PRICE"
	SymbolFilterTypeMarketLotSize    SymbolFilterType = "MARKET_LOT_SIZE"
	SymbolFilterTypeMaxNumOrders     SymbolFilterType = "MAX_NUM_ORDERS"
	SymbolFilterTypeMaxNumAlgoOrders SymbolFilterType = "MAX_NUM_ALGO_ORDERS"
	SymbolFilterTypeMinNotional      SymbolFilterType = "MIN_NOTIONAL"
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

	// PercentPriceFilter define percent price filter of symbol
	PercentPriceFilter struct {
		MultiplierDecimal string `json:"multiplierDecimal"`
		MultiplierUp      string `json:"multiplierUp"`
		MultiplierDown    string `json:"multiplierDown"`
	}

	// MarketLotSizeFilter define market lot size filter of symbol
	MarketLotSizeFilter struct {
		MaxQuantity string `json:"maxQty"`
		MinQuantity string `json:"minQty"`
		StepSize    string `json:"stepSize"`
	}

	// MaxNumOrdersFilter define max num orders filter of symbol
	MaxNumOrdersFilter struct {
		Limit int64 `json:"limit"`
	}

	// MaxNumAlgoOrdersFilter define max num algo orders filter of symbol
	MaxNumAlgoOrdersFilter struct {
		Limit int64 `json:"limit"`
	}

	// MinNotionalFilter define min notional filter of symbol
	MinNotionalFilter struct {
		Notional string `json:"notional"`
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
		if filter["filterType"].(string) == string(SymbolFilterTypePrice) {
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

// PercentPriceFilter return percent price filter of symbol
func (s *Symbol) PercentPriceFilter() *PercentPriceFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypePercentPrice) {
			f := &PercentPriceFilter{}
			if i, ok := filter["multiplierDecimal"]; ok {
				f.MultiplierDecimal = i.(string)
			}
			if i, ok := filter["multiplierUp"]; ok {
				f.MultiplierUp = i.(string)
			}
			if i, ok := filter["multiplierDown"]; ok {
				f.MultiplierDown = i.(string)
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

// MaxNumOrdersFilter return max num orders filter of symbol
func (s *Symbol) MaxNumOrdersFilter() *MaxNumOrdersFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMaxNumOrders) {
			f := &MaxNumOrdersFilter{}
			if i, ok := filter["limit"]; ok {
				if limit, okk := ToInt64(i); okk == nil {
					f.Limit = limit
				}
			}
			return f
		}
	}
	return nil
}

// MaxNumAlgoOrdersFilter return max num orders filter of symbol
func (s *Symbol) MaxNumAlgoOrdersFilter() *MaxNumAlgoOrdersFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMaxNumAlgoOrders) {
			f := &MaxNumAlgoOrdersFilter{}
			if i, ok := filter["limit"]; ok {
				if limit, okk := ToInt64(i); okk == nil {
					f.Limit = limit
				}
			}
			return f
		}
	}
	return nil
}

// MinNotionalFilter return min notional filter of symbol
func (s *Symbol) MinNotionalFilter() *MinNotionalFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == string(SymbolFilterTypeMinNotional) {
			f := &MinNotionalFilter{}
			if i, ok := filter["notional"]; ok {
				f.Notional = i.(string)
			}
			return f
		}
	}
	return nil
}

func ToInt64(digit interface{}) (i int64, err error) {
	if intVal, ok := digit.(int); ok {
		return int64(intVal), nil
	}
	if floatVal, ok := digit.(float64); ok {
		return int64(floatVal), nil
	}
	return 0, fmt.Errorf("unexpected digit: %v", digit)
}
