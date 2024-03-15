package filters

import (
	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	FilterName string
	Filter     map[string]interface{}
	FilterItem struct {
		FilterName
		Filter
	}
	Filters struct {
		btree.BTree
	}
)

func (s *FilterItem) Less(than btree.Item) bool {
	return s.FilterName < than.(*FilterItem).FilterName
}

func (s *FilterItem) Equal(than btree.Item) bool {
	return s.FilterName == than.(*FilterItem).FilterName
}

func NewFilters(degree int) *Filters {
	return &Filters{
		BTree: *btree.New(degree),
	}
}

func (s *Filters) Insert(filterName FilterName, filter Filter) {
	s.ReplaceOrInsert(&FilterItem{filterName, filter})
}

func (s *Filters) Init(filters []Filter) {
	for _, filter := range filters {
		filterName, ok := filter["filterType"].(FilterName)
		if ok {
			s.Insert(filterName, filter)
		}
	}
}

func (s *Filters) GetFilter(symbol string) *FilterItem {
	item := s.Get(&FilterItem{FilterName: FilterName(symbol)})
	if item == nil {
		return nil
	}
	return item.(*FilterItem)
}

func (s *Filters) DeleteFilter(symbol string) {
	s.Delete(&FilterItem{FilterName: FilterName(symbol)})
}

func (s *Filters) Len() int {
	return s.BTree.Len()
}

///////////////////////////////////////////////

func (s *Filters) GetFilters_Price_Filter() *binance.PriceFilter {
	filter := s.GetFilter("PRICE_FILTER")
	priceFilter := binance.PriceFilter{
		MinPrice: filter.Filter["minPrice"].(string),
		MaxPrice: filter.Filter["maxPrice"].(string),
		TickSize: filter.Filter["tickSize"].(string),
	}
	return &priceFilter
}

func (s *Filters) GetFilers_Lot_Size_Filter() *binance.LotSizeFilter {
	filter := s.GetFilter("LOT_SIZE")
	lotSizeFilter := binance.LotSizeFilter{
		MinQuantity: filter.Filter["minQty"].(string),
		MaxQuantity: filter.Filter["maxQty"].(string),
		StepSize:    filter.Filter["stepSize"].(string),
	}
	return &lotSizeFilter
}

func (s *Filters) GetFilters_Iceberg_Parts_Filter() *binance.IcebergPartsFilter {
	filter := s.GetFilter("ICEBERG_PARTS")
	icebergPartsFilter := binance.IcebergPartsFilter{
		Limit: int(filter.Filter["limit"].(float64)),
	}
	return &icebergPartsFilter
}

func (s *Filters) GetFilters_Market_Lot_Size_Filter() *binance.MarketLotSizeFilter {
	filter := s.GetFilter("MARKET_LOT_SIZE")
	marketLotSizeFilter := binance.MarketLotSizeFilter{
		MinQuantity: filter.Filter["minQty"].(string),
		MaxQuantity: filter.Filter["maxQty"].(string),
		StepSize:    filter.Filter["stepSize"].(string),
	}
	return &marketLotSizeFilter
}

type TrailingDelta struct {
	minTrailingAboveDelta float64
	minTrailingBelowDelta float64
	maxTrailingAboveDelta float64
	maxTrailingBelowDelta float64
}

func (s *Filters) GetFilters_Trailing_Delta_Filter() TrailingDelta {
	filter := s.GetFilter("TRAILING_STOP")
	trailingStopFilter := TrailingDelta{
		minTrailingAboveDelta: filter.Filter["minTrailingStop"].(float64),
		minTrailingBelowDelta: filter.Filter["minTrailingStop"].(float64),
		maxTrailingAboveDelta: filter.Filter["maxTrailingStop"].(float64),
		maxTrailingBelowDelta: filter.Filter["maxTrailingStop"].(float64),
	}
	return trailingStopFilter
}

type Percent_Price_By_Side struct {
	bidMultiplierUp   string
	bidMultiplierDown string
	askMultiplierUp   string
	askMultiplierDown string
	averagePriceMins  int
}

func (s *Filters) GetFilters_Percent_Price_By_Side_Filter() *Percent_Price_By_Side {
	filter := s.GetFilter("PERCENT_PRICE_BY_SIDE")
	percentPriceFilter := Percent_Price_By_Side{
		bidMultiplierUp:   filter.Filter["bidMultiplierUp"].(string),
		bidMultiplierDown: filter.Filter["bidMultiplierDown"].(string),
		askMultiplierUp:   filter.Filter["askMultiplierUp"].(string),
		askMultiplierDown: filter.Filter["askMultiplierDown"].(string),
		averagePriceMins:  int(filter.Filter["averagePriceMins"].(float64)),
	}
	return &percentPriceFilter
}

func (s *Filters) GetFilters_Notional_Filter() *binance.NotionalFilter {
	filter := s.GetFilter("NOTIONAL")
	notionalFilter := binance.NotionalFilter{
		MinNotional: filter.Filter["minNotional"].(string),
	}
	return &notionalFilter
}

func (s *Filters) GetFilters_Max_Num_Orders_Filter() float64 {
	filter := s.GetFilter("MAX_NUM_ORDERS")
	maxNumOrdersFilter := float64(filter.Filter["maxNumOrders"].(float64))
	return maxNumOrdersFilter
}

func (s *Filters) GetFilters_Max_Num_Algo_Orders_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.MaxNumAlgoOrdersFilter {
	filter := s.GetFilter("MAX_NUM_ALGO_ORDERS")
	maxNumAlgoOrdersFilter := binance.MaxNumAlgoOrdersFilter{
		MaxNumAlgoOrders: int(filter.Filter["maxNumAlgoOrders"].(float64)),
	}
	return &maxNumAlgoOrdersFilter
}
