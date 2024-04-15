package symbols

import (
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	symbol_info "github.com/fr0ster/go-trading-utils/types/info/symbols/symbol"
	"github.com/google/btree"
)

type (
	Symbols interface {
		Lock()
		Unlock()
		GetSymbol(symbol string) *symbol_info.Symbol
		GetSpotSymbol() *binance.Symbol
		GetFuturesSymbol() *futures.Symbol
		Insert(symbol btree.Item)
		Len() int
	}
)
