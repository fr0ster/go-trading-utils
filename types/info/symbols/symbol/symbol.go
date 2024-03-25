package symbol

import (
	"sync"

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
		mu                         sync.Mutex
	}
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

func (s *Symbol) Lock() {
	s.mu.Lock()
}

func (s *Symbol) TryLock() bool {
	return s.mu.TryLock()
}

func (s *Symbol) Unlock() {
	s.mu.Unlock()
}

func (s *Symbol) GetFilter(filterType string) interface{} {
	for _, filter := range s.Filters {
		if _, exists := filter["filterType"]; exists && filter["filterType"] == filterType {
			return &filter
		}
	}
	return nil
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
