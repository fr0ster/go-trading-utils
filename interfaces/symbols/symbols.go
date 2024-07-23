package symbols

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/google/btree"
)

type (
	Symbols interface {
		Lock()
		Unlock()
		GetSymbol(symbol string) *symbol_info.SymbolInfo
		GetSpotSymbol() *binance.Symbol
		GetFuturesSymbol() *futures.Symbol
		Ascend(f func(btree.Item) bool)
		Descend(f func(btree.Item) bool)
		Insert(symbol btree.Item)
		Len() int
	}
)
