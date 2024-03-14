package filters

import (
	"sync"

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
		mu sync.Mutex
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
		mu:    sync.Mutex{},
	}
}

func (s *Filters) Lock() {
	s.mu.Lock()
}

func (s *Filters) Unlock() {
	s.mu.Unlock()
}

func (s *Filters) Insert(filterName FilterName, filter Filter) {
	s.ReplaceOrInsert(&FilterItem{filterName, filter})
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

func (s *Filters) Init(filters []Filter) {
	for _, filter := range filters {
		s.Insert(filter["filterType"].(FilterName), filter)
	}
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

// func GetFilers_Lot_Size_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.LotSizeFilter {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "LOT_SIZE" {
// 					lotSizeFilter := binance.LotSizeFilter{
// 						MinQuantity: filter["minQty"].(string),
// 						MaxQuantity: filter["maxQty"].(string),
// 						StepSize:    filter["stepSize"].(string),
// 					}
// 					return &lotSizeFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func GetFilters_Iceberg_Parts_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.IcebergPartsFilter {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "ICEBERG_PARTS" {
// 					icebergPartsFilter := binance.IcebergPartsFilter{
// 						Limit: int(filter["limit"].(float64)),
// 					}
// 					return &icebergPartsFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func GetFilters_Market_Lot_Size_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.MarketLotSizeFilter {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "MARKET_LOT_SIZE" {
// 					marketLotSizeFilter := binance.MarketLotSizeFilter{
// 						MinQuantity: filter["minQty"].(string),
// 						MaxQuantity: filter["maxQty"].(string),
// 						StepSize:    filter["stepSize"].(string),
// 					}
// 					return &marketLotSizeFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// type TrailingDelta struct {
// 	minTrailingAboveDelta float64
// 	minTrailingBelowDelta float64
// 	maxTrailingAboveDelta float64
// 	maxTrailingBelowDelta float64
// }

// func GetFilters_Trailing_Delta_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) TrailingDelta {
// 	trailingStopFilter := TrailingDelta{}
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "TRAILING_STOP" {
// 					trailingStopFilter = TrailingDelta{
// 						minTrailingAboveDelta: filter["minTrailingStop"].(float64),
// 						minTrailingBelowDelta: filter["minTrailingStop"].(float64),
// 						maxTrailingAboveDelta: filter["maxTrailingStop"].(float64),
// 						maxTrailingBelowDelta: filter["maxTrailingStop"].(float64),
// 					}
// 					return trailingStopFilter
// 				}
// 			}
// 		}
// 	}
// 	return TrailingDelta{}
// }

// type Percent_Price_By_Side struct {
// 	bidMultiplierUp   string
// 	bidMultiplierDown string
// 	askMultiplierUp   string
// 	askMultiplierDown string
// 	averagePriceMins  int
// }

// func GetFilters_Percent_Price_By_Side_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string, side string) *Percent_Price_By_Side {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "PERCENT_PRICE_BY_SIDE" {
// 					percentPriceFilter := Percent_Price_By_Side{
// 						bidMultiplierUp:   filter["bidMultiplierUp"].(string),
// 						bidMultiplierDown: filter["bidMultiplierDown"].(string),
// 						askMultiplierUp:   filter["askMultiplierUp"].(string),
// 						askMultiplierDown: filter["askMultiplierDown"].(string),
// 						averagePriceMins:  int(filter["averagePriceMins"].(float64)),
// 					}
// 					return &percentPriceFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func GetFilters_Notional_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.NotionalFilter {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "NOTIONAL" {
// 					notionalFilter := binance.NotionalFilter{
// 						MinNotional: filter["minNotional"].(string),
// 					}
// 					return &notionalFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func GetFilters_Max_Num_Orders_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) float64 {
// 	maxNumOrdersFilter := 0.0
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "MAX_NUM_ORDERS" {
// 					maxNumOrdersFilter = float64(filter["maxNumOrders"].(float64))
// 					return maxNumOrdersFilter
// 				}
// 			}
// 		}
// 	}
// 	return 0
// }

// func GetFilters_Max_Num_Algo_Orders_Filter(exchangeInfo *binance.ExchangeInfo, symbolname string) *binance.MaxNumAlgoOrdersFilter {
// 	for _, info := range exchangeInfo.Symbols {
// 		if info.Symbol == symbolname {
// 			for _, filter := range info.Filters {
// 				if filter["filterType"] == "MAX_NUM_ALGO_ORDERS" {
// 					maxNumAlgoOrdersFilter := binance.MaxNumAlgoOrdersFilter{
// 						MaxNumAlgoOrders: int(filter["maxNumAlgoOrders"].(float64)),
// 					}
// 					return &maxNumAlgoOrdersFilter
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
