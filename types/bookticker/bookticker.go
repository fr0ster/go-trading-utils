package bookticker

import (
	"sync"

	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	BookTickerItem struct {
		Symbol      string
		BidPrice    float64
		BidQuantity float64
		AskPrice    float64
		AskQuantity float64
	}
	BookTickerBTree struct {
		tree   *btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *BookTickerItem) Less(than btree.Item) bool {
	return i.Symbol < than.(*BookTickerItem).Symbol
}

func (i *BookTickerItem) Equal(than btree.Item) bool {
	return i.Symbol == than.(*BookTickerItem).Symbol
}

func (btt *BookTickerBTree) Lock() {
	btt.mutex.Lock()
}

func (btt *BookTickerBTree) Unlock() {
	btt.mutex.Unlock()
}

func (btt *BookTickerBTree) Ascend(f func(item btree.Item) bool) {
	btt.tree.Ascend(f)
}

func (btt *BookTickerBTree) Descend(f func(item btree.Item) bool) {
	btt.tree.Descend(f)
}

func (btt *BookTickerBTree) Get(symbol string) btree.Item {
	item := btt.tree.Get(&BookTickerItem{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item
}

func (btt *BookTickerBTree) Set(item btree.Item) {
	btt.tree.ReplaceOrInsert(item)
}

func New(degree int) *BookTickerBTree {
	return &BookTickerBTree{
		tree:   btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func Binance2BookTicker(binanceBookTicker interface{}) (*BookTickerItem, error) {
	var bookTickerItem BookTickerItem
	err := copier.Copy(&bookTickerItem, binanceBookTicker)
	if err != nil {
		return nil, err
	}
	return &bookTickerItem, nil
}
