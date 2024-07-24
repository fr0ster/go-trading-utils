package symbols

import (
	"sync"

	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/google/btree"
)

type (
	Symbols struct {
		degree     int
		symbols    btree.BTree
		mu         sync.Mutex
		getSymbols func() []*symbol_info.Symbol
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

func (s *Symbols) Ascend(f func(btree.Item) bool) {
	s.symbols.Ascend(func(i btree.Item) bool {
		return f(i)
	})
}

func (s *Symbols) Descend(f func(btree.Item) bool) {
	s.symbols.Descend(func(i btree.Item) bool {
		return f(i)
	})
}

func New(
	degree int,
	getSymbols func() []*symbol_info.Symbol) (s *Symbols, err error) {
	s = &Symbols{
		degree:  degree,
		symbols: *btree.New(degree),
		mu:      sync.Mutex{},
	}
	if getSymbols != nil {
		for _, symbol := range getSymbols() {
			s.Insert(symbol)
		}
	}
	return
}
