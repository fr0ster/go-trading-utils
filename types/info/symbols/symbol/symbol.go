package symbol

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Symbol struct {
		Symbol                     string                   `json:"symbol"`
		Status                     string                   `json:"status"`
		BaseAsset                  string                   `json:"baseAsset"`
		BaseAssetPrecision         int                      `json:"baseAssetPrecision"`
		QuoteAsset                 string                   `json:"quoteAsset"`
		QuotePrecision             int                      `json:"quotePrecision"`
		QuoteAssetPrecision        int                      `json:"quoteAssetPrecision"`
		BaseCommissionPrecision    int32                    `json:"baseCommissionPrecision"`
		QuoteCommissionPrecision   int32                    `json:"quoteCommissionPrecision"`
		OrderTypes                 []string                 `json:"orderTypes"`
		IcebergAllowed             bool                     `json:"icebergAllowed"`
		OcoAllowed                 bool                     `json:"ocoAllowed"`
		QuoteOrderQtyMarketAllowed bool                     `json:"quoteOrderQtyMarketAllowed"`
		IsSpotTradingAllowed       bool                     `json:"isSpotTradingAllowed"`
		IsMarginTradingAllowed     bool                     `json:"isMarginTradingAllowed"`
		Filters                    []map[string]interface{} `json:"filters"`
		Permissions                []string                 `json:"permissions"`
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

	// The "Algo" order is STOP_LOSS, STOP_LOS_LIMITED, TAKE_PROFIT and TAKE_PROFIT_Limit Stop Loss Order.
	// Therefore, orders other than the above types are non conditional(Algo) orders, and MaxNumOrders defines the maximum
	// number of orders placed for these types of orders
	MaxNumOrdersFilter struct {
		MaxNumOrders int `json:"maxNumOrders"`
	}

	// MaxNumAlgoOrdersFilter define max num algo orders filter of symbol
	MaxNumAlgoOrdersFilter struct {
		MaxNumAlgoOrders int `json:"maxNumAlgoOrders"`
	}
	SymbolFilterType string
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

func (s *Symbol) Less(than btree.Item) bool {
	return s.Symbol < than.(*Symbol).Symbol
}

func (s *Symbol) Equal(than btree.Item) bool {
	return s.Symbol == than.(*Symbol).Symbol
}

func (s *Symbol) GetSymbol() string {
	return s.Symbol
}

func (s *Symbol) GetFilter(filterType string) interface{} {
	for _, filter := range s.Filters {
		if _, exists := filter["filterType"]; exists && filter["filterType"] == filterType {
			return &filter
		}
	}
	return nil
}

func (s *Symbol) GetSpotSymbol() (*binance.Symbol, error) {
	var outSymbol binance.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

func (s *Symbol) GetFuturesSymbol() (*futures.Symbol, error) {
	var outSymbol futures.Symbol
	err := copier.Copy(&outSymbol, s)
	if err != nil {
		return nil, err
	}
	return &outSymbol, nil
}

func NewSymbol(symbol interface{}) *Symbol {
	val, _ := Binance2Symbol(symbol)
	return val
}

func Binance2Symbol(binanceSymbol interface{}) (*Symbol, error) {
	var symbol Symbol
	err := copier.Copy(&symbol, binanceSymbol)
	if err != nil {
		return nil, err
	}
	return &symbol, nil
}
