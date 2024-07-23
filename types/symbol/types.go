package symbol

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/google/btree"
)

type (
	OrderType  string
	QuoteAsset string
	BaseAsset  string
	// Дані про обмеження на пару
	SymbolInfo struct {
		Symbol                 string
		notional               items_types.ValueType
		stepSize               items_types.QuantityType
		maxQty                 items_types.QuantityType
		minQty                 items_types.QuantityType
		tickSize               items_types.PriceType
		maxPrice               items_types.PriceType
		minPrice               items_types.PriceType
		baseSymbol             QuoteAsset
		targetSymbol           BaseAsset
		isMarginTradingAllowed bool
		permissions            []string
		orderType              []OrderType
	}
)

func (si *SymbolInfo) Less(than btree.Item) bool {
	return si.Symbol < than.(*SymbolInfo).Symbol
}

func (si *SymbolInfo) Equal(than btree.Item) bool {
	return si.Symbol == than.(*SymbolInfo).Symbol
}

func (si *SymbolInfo) GetSymbol() string {
	return si.Symbol
}

func (si *SymbolInfo) GetNotional() items_types.ValueType {
	return si.notional
}

func (si *SymbolInfo) GetStepSize() items_types.QuantityType {
	return si.stepSize
}

func (si *SymbolInfo) GetMaxQty() items_types.QuantityType {
	return si.maxQty
}

func (si *SymbolInfo) GetMinQty() items_types.QuantityType {
	return si.minQty
}

func (si *SymbolInfo) GetTickSizeExp() items_types.PriceType {
	return si.tickSize
}

func (si *SymbolInfo) GetMaxPrice() items_types.PriceType {
	return si.maxPrice
}

func (si *SymbolInfo) GetMinPrice() items_types.PriceType {
	return si.minPrice
}

func (si *SymbolInfo) GetBaseSymbol() QuoteAsset {
	return si.baseSymbol
}

func (si *SymbolInfo) GetTargetSymbol() BaseAsset {
	return si.targetSymbol
}

func (si *SymbolInfo) IsMarginTradingAllowed() bool {
	return si.isMarginTradingAllowed
}

func (si *SymbolInfo) GetPermissions() []string {
	return si.permissions
}

func (si *SymbolInfo) GetOrderType() []OrderType {
	return si.orderType
}

func New(
	symbol string,
	notional items_types.ValueType,
	stepSize items_types.QuantityType,
	maxQty items_types.QuantityType,
	minQty items_types.QuantityType,
	tickSize items_types.PriceType,
	maxPrice items_types.PriceType,
	minPrice items_types.PriceType,
	quoteAsset QuoteAsset,
	baseAsset BaseAsset,
	isMarginTradingAllowed bool,
	permissions []string,
	orderType []OrderType) *SymbolInfo {
	return &SymbolInfo{
		Symbol:                 symbol,
		notional:               notional,
		stepSize:               stepSize,
		maxQty:                 maxQty,
		minQty:                 minQty,
		tickSize:               tickSize,
		maxPrice:               maxPrice,
		minPrice:               minPrice,
		baseSymbol:             quoteAsset,
		targetSymbol:           baseAsset,
		isMarginTradingAllowed: isMarginTradingAllowed,
		permissions:            permissions,
		orderType:              orderType,
	}
}
