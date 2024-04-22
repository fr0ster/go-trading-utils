package symbols

import (
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/google/btree"
)

type (
	Symbols struct {
		degree  int
		symbols btree.BTree
		mu      sync.Mutex
	}
	ISymbol interface {
		symbol_info.SpotSymbol | symbol_info.FuturesSymbol
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

func (s *Symbols) Insert(symbol btree.Item) {
	s.symbols.ReplaceOrInsert(symbol)
}

func (s *Symbols) GetSymbol(symbol btree.Item) btree.Item {
	item := s.symbols.Get(symbol)
	if item == nil {
		return nil
	}
	return item
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

func NewSymbols(degree int, symbols []interface{}) (s *Symbols, err error) {
	s = &Symbols{
		degree:  degree,
		symbols: *btree.New(degree),
		mu:      sync.Mutex{},
	}
	for _, symbol := range symbols {
		switch symbol := symbol.(type) {
		case symbol_info.SpotSymbol:
			s.Insert(&symbol)
		case binance.Symbol:
			val := symbol_info.SpotSymbol(symbol)
			s.Insert(&val)
		case symbol_info.FuturesSymbol:
			s.Insert(&symbol)
		case futures.Symbol:
			val := symbol_info.FuturesSymbol(symbol)
			s.Insert(&val)
		default:
			err = fmt.Errorf("invalid symbol type: %T", symbol)
		}
	}
	return
}
