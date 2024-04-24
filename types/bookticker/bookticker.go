package bookticker

import (
	"sync"

	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	BookTicker struct {
		Symbol      string
		BidPrice    float64
		BidQuantity float64
		AskPrice    float64
		AskQuantity float64
	}
	BookTickers struct {
		tree   *btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *BookTicker) Less(than btree.Item) bool {
	return i.Symbol < than.(*BookTicker).Symbol
}

func (i *BookTicker) Equal(than btree.Item) bool {
	return i.Symbol == than.(*BookTicker).Symbol
}

func (btt *BookTickers) Lock() {
	btt.mutex.Lock()
}

func (btt *BookTickers) Unlock() {
	btt.mutex.Unlock()
}

func (btt *BookTickers) Ascend(f func(item btree.Item) bool) {
	btt.tree.Ascend(f)
}

func (btt *BookTickers) Descend(f func(item btree.Item) bool) {
	btt.tree.Descend(f)
}

func (btt *BookTickers) Get(symbol string) btree.Item {
	item := btt.tree.Get(&BookTicker{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item
}

func (btt *BookTickers) Set(item btree.Item) {
	btt.tree.ReplaceOrInsert(item)
}

func New(degree int) *BookTickers {
	return &BookTickers{
		tree:   btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func Binance2BookTicker(binanceBookTicker interface{}) (*BookTicker, error) {
	var bookTickerItem BookTicker
	err := copier.Copy(&bookTickerItem, binanceBookTicker)
	if err != nil {
		return nil, err
	}
	return &bookTickerItem, nil
}
