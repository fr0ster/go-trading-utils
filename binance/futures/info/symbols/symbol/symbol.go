package symbol

import (
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
)

type (
	SymbolName string
	Symbol     struct {
		SymbolName
		futures.Symbol
		mu sync.Mutex
	}
)

func (s *Symbol) Less(than btree.Item) bool {
	return s.SymbolName < than.(*Symbol).SymbolName
}

func (s *Symbol) Equal(than btree.Item) bool {
	return s.SymbolName == than.(*Symbol).SymbolName
}

func NewSymbol(degree int, symbol *futures.Symbol) *Symbol {
	return &Symbol{
		SymbolName: SymbolName(symbol.Symbol),
		Symbol:     *symbol,
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
