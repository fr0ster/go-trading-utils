package symbols

import (
	"sync"

	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/google/btree"
)

type (
	Symbols struct {
		degree  int
		symbols btree.BTree
		mu      sync.Mutex
	}
)

func (s *Symbols) Lock() {
	s.mu.Lock()
}

func (s *Symbols) Unlock() {
	s.mu.Unlock()
}

func (s *Symbols) Len() int {
	return s.symbols.Len()
}

func (s *Symbols) Insert(symbol *symbol_info.Symbol) {
	s.symbols.ReplaceOrInsert(symbol)
}

func (s *Symbols) GetSymbol(symbol string) *symbol_info.Symbol {
	item := s.symbols.Get(&symbol_info.Symbol{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*symbol_info.Symbol)
}

func NewSymbols(degree int, symbols []interface{}) *Symbols {
	s := Symbols{
		degree:  degree,
		symbols: *btree.New(degree),
		mu:      sync.Mutex{},
	}
	for _, symbol := range symbols {
		s.Insert(symbol_info.NewSymbol(symbol))
	}
	return &s
}
