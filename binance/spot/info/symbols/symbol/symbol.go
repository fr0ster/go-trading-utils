package symbol

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type (
	SymbolName string
	Symbol     struct {
		SymbolName
		binance.Symbol
		mu sync.Mutex
	}
)

func (s *Symbol) Less(than btree.Item) bool {
	return s.SymbolName < than.(*Symbol).SymbolName
}

func (s *Symbol) Equal(than btree.Item) bool {
	return s.SymbolName == than.(*Symbol).SymbolName
}

func (s *Symbol) GetSymbol() string {
	return s.Symbol.Symbol
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

func NewSymbol(degree int, symbol *binance.Symbol) *Symbol {
	return &Symbol{
		SymbolName: SymbolName(symbol.Symbol),
		Symbol:     *symbol,
		mu:         sync.Mutex{},
	}
}
