package bookticker

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	bookticker_interface "github.com/fr0ster/go-trading-utils/interfaces/bookticker"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	BookTickerBTree struct {
		tree   *btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

func New(degree int) *BookTickerBTree {
	return &BookTickerBTree{
		tree:   btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

func (btt *BookTickerBTree) Lock() {
	btt.mutex.Lock()
}

func (btt *BookTickerBTree) Unlock() {
	btt.mutex.Unlock()
}

func (btt *BookTickerBTree) Init(api_key, secret_key, symbolname string, UseTestnet bool) (err error) {
	binance.UseTestnet = UseTestnet
	client := binance.NewClient(api_key, secret_key)
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	for _, bookTicker := range bookTickerList {
		btt.tree.ReplaceOrInsert(&bookticker_interface.BookTickerItem{
			Symbol:      bookTicker.Symbol,
			BidPrice:    utils.ConvStrToFloat64(bookTicker.BidPrice),
			BidQuantity: utils.ConvStrToFloat64(bookTicker.BidQuantity),
			AskPrice:    utils.ConvStrToFloat64(bookTicker.AskPrice),
			AskQuantity: utils.ConvStrToFloat64(bookTicker.AskQuantity),
		})
	}
	return nil
}

func (btt *BookTickerBTree) Ascend(f func(item btree.Item) bool) {
	btt.tree.Ascend(f)
}

func (btt *BookTickerBTree) Descend(f func(item btree.Item) bool) {
	btt.tree.Descend(f)
}

func (btt *BookTickerBTree) Get(symbol string) *bookticker_interface.BookTickerItem {
	item := btt.tree.Get(&bookticker_interface.BookTickerItem{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*bookticker_interface.BookTickerItem)
}

func (btt *BookTickerBTree) Set(item bookticker_interface.BookTickerItem) {
	btt.tree.ReplaceOrInsert(&item)
}
