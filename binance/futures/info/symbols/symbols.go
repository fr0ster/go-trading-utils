package symbols

import (
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	symbol_info "github.com/fr0ster/go-trading-utils/binance/futures/info/symbols/symbol"
	"github.com/google/btree"
)

type (
	Symbols struct {
		degree int
		btree.BTree
		mu sync.Mutex
	}
)

func NewSymbols(degree int, symbols []futures.Symbol) *Symbols {
	s := Symbols{
		degree: degree,
		BTree:  *btree.New(degree),
		mu:     sync.Mutex{},
	}
	for _, symbol := range symbols {
		s.Insert((*symbol_info.Symbol)(symbol_info.NewSymbol(s.degree, &symbol)))
	}
	return &s
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

func (s *Symbols) Insert(symbol *symbol_info.Symbol) {
	s.ReplaceOrInsert(symbol)
}

func (s *Symbols) GetSymbol(symbol string) *symbol_info.Symbol {
	item := s.Get(&symbol_info.Symbol{SymbolName: symbol_info.SymbolName(symbol)})
	if item == nil {
		return nil
	}
	return item.(*symbol_info.Symbol)
}
