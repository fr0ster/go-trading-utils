package symbol

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	filters_info "github.com/fr0ster/go-binance-utils/spot/info/symbols/filters"
	"github.com/google/btree"
)

type (
	SymbolName string
	Symbol     struct {
		SymbolName
		binance.Symbol
		filters_info.Filters
		mu sync.Mutex
	}
)

func (s *Symbol) Less(than btree.Item) bool {
	return s.SymbolName < than.(*Symbol).SymbolName
}

func (s *Symbol) Equal(than btree.Item) bool {
	return s.SymbolName == than.(*Symbol).SymbolName
}

func NewSymbol(degree int, symbol *binance.Symbol) *Symbol {
	filters := filters_info.NewFilters(degree)
	filters.Init(convertFilters(symbol.Filters))
	return &Symbol{
		SymbolName: SymbolName(symbol.Symbol),
		Symbol:     *symbol,
		Filters:    *filters,
		mu:         sync.Mutex{},
	}
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

func convertFilters(filters []map[string]interface{}) []filters_info.Filter {
	convertedFilters := make([]filters_info.Filter, len(filters))
	for i, f := range filters {
		convertedFilters[i] = filters_info.Filter(f)
	}
	return convertedFilters
}
