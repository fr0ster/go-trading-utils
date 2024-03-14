package info

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	Symbol  binance.Symbol
	Symbols struct {
		btree.BTree
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
		BTree: *btree.New(degree),
		mu:    sync.Mutex{},
	}
}

func (s *Symbols) Insert(symbol *Symbol) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReplaceOrInsert(symbol)
}

func (s *Symbols) GetSymbol(symbol string) *Symbol {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.Get(&Symbol{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*Symbol)
}

func (s *Symbols) DeleteSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Delete(&Symbol{Symbol: symbol})
}

func (s *Symbols) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BTree.Len()
}

func (s *Symbols) Init(symbols []binance.Symbol) error {
	for _, symbol := range symbols {
		s.Insert((*Symbol)(&symbol))
	}
	return nil
}
