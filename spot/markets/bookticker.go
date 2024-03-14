package markets

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
)

type (
	BookTickerBTree struct {
		*btree.BTree
		sync.Mutex
	}
	BookTickerItemType struct {
		Symbol      SymbolType
		BidPrice    PriceType
		BidQuantity PriceType
		AskPrice    PriceType
		AskQuantity PriceType
	}
	PriceType  float64
	SymbolType string
)

// var tree.Mutex sync.Mutex

func BookTickerNew(degree int) *BookTickerBTree {
	return &BookTickerBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
}

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (i *BookTickerItemType) Less(than btree.Item) bool {
	return i.Symbol < than.(*BookTickerItemType).Symbol
}

func (i *BookTickerItemType) Equal(than btree.Item) bool {
	return i.Symbol == than.(*BookTickerItemType).Symbol
}

// Init initializes the book ticker tree with prices.
// It retrieves the book tickers for the given symbol from the Binance client
// and inserts them into the book ticker tree.
func (tree *BookTickerBTree) Init(client *binance.Client, symbolname string) (err error) {
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	for _, bookTicker := range bookTickerList {
		tree.ReplaceOrInsert(&BookTickerItemType{
			Symbol:      SymbolType(bookTicker.Symbol),
			BidPrice:    PriceType(utils.ConvStrToFloat64(bookTicker.BidPrice)),
			BidQuantity: PriceType(utils.ConvStrToFloat64(bookTicker.BidQuantity)),
			AskPrice:    PriceType(utils.ConvStrToFloat64(bookTicker.AskPrice)),
			AskQuantity: PriceType(utils.ConvStrToFloat64(bookTicker.AskQuantity)),
		})
	}
	return nil
}

func (tree *BookTickerBTree) GetItem(symbol SymbolType) *BookTickerItemType {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	item := tree.Get(&BookTickerItemType{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*BookTickerItemType)
}

func (tree *BookTickerBTree) SetItem(item BookTickerItemType) {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	tree.ReplaceOrInsert(&item)
}

func (tree *BookTickerBTree) GetBySymbol(symbol SymbolType) *BookTickerBTree {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	newTree := BookTickerNew(2)
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*BookTickerItemType)
		if item.Symbol == symbol {
			newTree.ReplaceOrInsert(item)
		}
		return true
	})
	return newTree
}

func (tree *BookTickerBTree) GetByBidPrice(symbol SymbolType, bidPrice PriceType) *BookTickerBTree {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	newTree := BookTickerNew(2)
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*BookTickerItemType)
		if item.Symbol == symbol && item.BidPrice == bidPrice {
			newTree.ReplaceOrInsert(item)
		}
		return true
	})
	return newTree
}

func (tree *BookTickerBTree) GetByAskPrice(symbol SymbolType, askPrice PriceType) *BookTickerBTree {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	newTree := BookTickerNew(2)
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*BookTickerItemType)
		if item.Symbol == symbol && item.AskPrice == askPrice {
			newTree.ReplaceOrInsert(item)
		}
		return true
	})
	return newTree
}

// Show prints the book ticker information for each item in the BookTickerTree.
func (tree *BookTickerBTree) Show() {
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*BookTickerItemType)
		println(
			"Symbol:", item.Symbol,
			"BidPrice:", utils.ConvFloat64ToStr(float64(item.BidPrice), 8),
			"BidQuantity:", utils.ConvFloat64ToStr(float64(item.BidQuantity), 8),
			"AskPrice:", utils.ConvFloat64ToStr(float64(item.AskPrice), 8),
			"AskQuantity:", utils.ConvFloat64ToStr(float64(item.AskQuantity), 8),
		)
		return true
	})
}
