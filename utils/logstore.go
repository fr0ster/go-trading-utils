package utils

import (
	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type LogItem struct {
	Timestamp         int64
	AccountType       binance.AccountType
	Symbol            binance.SymbolType
	Balance           float64
	CalculatedBalance float64
	Quantity          float64
	Value             float64
	BoundQuantity     float64
	Msg               string
}

// DataStore represents the data store for your program
type (
	LogStore struct {
		FilePath string
	}
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b LogItem) Less(than btree.Item) bool {
	return b.Symbol < than.(LogItem).Symbol
}
