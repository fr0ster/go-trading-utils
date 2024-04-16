package symbol

type (
	FilterType string
	Symbol     interface {
		GetSymbolName() string
		GetFilter(filterType FilterType) interface{}
	}
)

const (
	// map[filterType:PRICE_FILTER maxPrice:1000000.00000000 minPrice:0.01000000 tickSize:0.01000000]
	// map[filterType:LOT_SIZE maxQty:9000.00000000 minQty:0.00001000 stepSize:0.00001000]
	// map[filterType:ICEBERG_PARTS limit:10]
	// map[filterType:MARKET_LOT_SIZE maxQty:93.14080512 minQty:0.00000000 stepSize:0.00000000]
	// map[filterType:TRAILING_DELTA maxTrailingAboveDelta:2000 maxTrailingBelowDelta:2000 minTrailingAboveDelta:10 minTrailingBelowDelta:10]
	// map[askMultiplierDown:0.2 askMultiplierUp:5 avgPriceMins:5 bidMultiplierDown:0.2 bidMultiplierUp:5 filterType:PERCENT_PRICE_BY_SIDE]
	// map[applyMaxToMarket:false applyMinToMarket:true avgPriceMins:5 filterType:NOTIONAL maxNotional:9000000.00000000 minNotional:5.00000000]
	// map[filterType:MAX_NUM_ORDERS maxNumOrders:200]
	// map[filterType:MAX_NUM_ALGO_ORDERS maxNumAlgoOrders:5]
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
