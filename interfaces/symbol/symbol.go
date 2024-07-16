package symbol

type (
	FilterType string
	Symbol     interface {
		GetSymbolName() string
		GetFilter(filterType FilterType) interface{}
	}
)

const (
	FilterTypePriceFilter        FilterType = "PRICE_FILTER"
	FilterTypeLotSize            FilterType = "LOT_SIZE"
	FilterTypeIceberg            FilterType = "ICEBERG_PARTS"
	FilterTypeMarketLotSize      FilterType = "MARKET_LOT_SIZE"
	FilterTypeTrailingDelta      FilterType = "TRAILING_DELTA"
	FilterTypePercentPriceBySide FilterType = "PERCENT_PRICE_BY_SIDE"
	FilterTypeNotional           FilterType = "NOTIONAL"
	FilterTypeMaxNumOrders       FilterType = "MAX_NUM_ORDERS"
	FilterTypeMaxNumAlgoOrders   FilterType = "MAX_NUM_ALGO_ORDERS"
)
