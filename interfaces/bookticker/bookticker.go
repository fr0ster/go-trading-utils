package bookticker

import (
	"github.com/google/btree"
)

type (
	BookTicker interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error)
		Ascend(f func(item btree.Item) bool)
		Descend(f func(item btree.Item) bool)
		Get(symbol string) *BookTickerItem
		Set(value BookTickerItem)
	}
	BookTickerItem struct {
		Symbol      string
		BidPrice    float64
		BidQuantity float64
		AskPrice    float64
		AskQuantity float64
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *BookTickerItem) Less(than btree.Item) bool {
	return i.Symbol < than.(*BookTickerItem).Symbol
}

func (i *BookTickerItem) Equal(than btree.Item) bool {
	return i.Symbol == than.(*BookTickerItem).Symbol
}
