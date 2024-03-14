package symbols

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	filters_info "github.com/fr0ster/go-binance-utils/spot/info/symbols/filters"
	"github.com/google/btree"
)

type (
	Symbol  binance.Symbol
	Symbols struct {
		btree.BTree
		filters_info.Filters
		mu sync.Mutex
	}
)

func (s *Symbol) Less(than btree.Item) bool {
	return s.Symbol < than.(*Symbol).Symbol
}

func (s *Symbol) Equal(than btree.Item) bool {
	return s.Symbol == than.(*Symbol).Symbol
}

func NewSymbols(degree int) *Symbols {
	return &Symbols{
		BTree:   *btree.New(degree),
		Filters: *filters_info.NewFilters(degree),
		mu:      sync.Mutex{},
	}
}

func (s *Symbols) Lock() {
	s.mu.Lock()
}

func (s *Symbols) Unlock() {
	s.mu.Unlock()
}

func (s *Symbols) Len() int {
	return s.BTree.Len()
}

func (s *Symbols) Insert(symbol *Symbol) {
	s.ReplaceOrInsert(symbol)
}

func (s *Symbols) GetSymbol(symbol string) *Symbol {
	item := s.Get(&Symbol{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*Symbol)
}

func (s *Symbols) DeleteSymbol(symbol string) {
	s.Delete(&Symbol{Symbol: symbol})
}

func (s *Symbols) Init(symbols []binance.Symbol) error {
	for _, symbol := range symbols {
		s.Insert((*Symbol)(&symbol))
		filterSlice := make([]filters_info.Filter, len(symbol.Filters))
		for i, f := range symbol.Filters {
			filterSlice[i] = filters_info.Filter(f)
		}
		s.Filters.Init(filterSlice)
	}
	return nil
}
