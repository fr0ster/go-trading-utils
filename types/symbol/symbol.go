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
	Symbol struct {
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

func (si *Symbol) Less(than btree.Item) bool {
	return si.Symbol < than.(*Symbol).Symbol
}

func (si *Symbol) Equal(than btree.Item) bool {
	return si.Symbol == than.(*Symbol).Symbol
}

func (si *Symbol) GetSymbol() string {
	return si.Symbol
}

func (si *Symbol) GetNotional() items_types.ValueType {
	return si.notional
}

func (si *Symbol) GetStepSize() items_types.QuantityType {
	return si.stepSize
}

func (si *Symbol) GetMaxQty() items_types.QuantityType {
	return si.maxQty
}

func (si *Symbol) GetMinQty() items_types.QuantityType {
	return si.minQty
}

func (si *Symbol) GetTickSizeExp() items_types.PriceType {
	return si.tickSize
}

func (si *Symbol) GetMaxPrice() items_types.PriceType {
	return si.maxPrice
}

func (si *Symbol) GetMinPrice() items_types.PriceType {
	return si.minPrice
}

func (si *Symbol) GetBaseSymbol() QuoteAsset {
	return si.baseSymbol
}

func (si *Symbol) GetTargetSymbol() BaseAsset {
	return si.targetSymbol
}

func (si *Symbol) IsMarginTradingAllowed() bool {
	return si.isMarginTradingAllowed
}

func (si *Symbol) GetPermissions() []string {
	return si.permissions
}

func (si *Symbol) GetOrderType() []OrderType {
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
	orderType []OrderType) *Symbol {
	return &Symbol{
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
