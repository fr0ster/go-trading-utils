package symbol_info

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

type (
	// Дані про обмеження на пару
	SymbolInfo struct {
		notional items_types.ValueType
		StepSize items_types.QuantityType
		maxQty   items_types.QuantityType
		minQty   items_types.QuantityType
		tickSize items_types.PriceType
		maxPrice items_types.PriceType
		minPrice items_types.PriceType
	}
)

func (si *SymbolInfo) GetNotional() items_types.ValueType {
	return si.notional
}

func (si *SymbolInfo) GetStepSize() items_types.QuantityType {
	return si.StepSize
}

func (si *SymbolInfo) GetMaxQty() items_types.QuantityType {
	return si.maxQty
}

func (si *SymbolInfo) GetMinQty() items_types.QuantityType {
	return si.minQty
}

func (si *SymbolInfo) GetTickSizeExp() int {
	return int(si.tickSize)
}

func (si *SymbolInfo) GetMaxPrice() items_types.PriceType {
	return si.maxPrice
}

func (si *SymbolInfo) GetMinPrice() items_types.PriceType {
	return si.minPrice
}

func New(
	notional items_types.ValueType,
	StepSize items_types.QuantityType,
	maxQty items_types.QuantityType,
	minQty items_types.QuantityType,
	tickSize items_types.PriceType,
	maxPrice items_types.PriceType,
	minPrice items_types.PriceType) *SymbolInfo {
	return &SymbolInfo{
		notional: notional,
		StepSize: StepSize,
		maxQty:   maxQty,
		minQty:   minQty,
		tickSize: tickSize,
		maxPrice: maxPrice,
		minPrice: minPrice,
	}
}
