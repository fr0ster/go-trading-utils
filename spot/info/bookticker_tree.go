package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
)

var (
	BookTickerTree     = btree.New(2) // Book ticker tree
	mu_bookticker_tree sync.Mutex     // Mutex for book ticker tree
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b BookTickerItem) Less(than btree.Item) bool {
	return b.Symbol < than.(BookTickerItem).Symbol
}

// InitPricesTree initializes the book ticker tree with prices.
// It retrieves the book tickers for the given symbol from the Binance client
// and inserts them into the book ticker tree.
func InitPricesTree(client *binance.Client, symbolname string) (err error) {
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	for _, bookTicker := range bookTickerList {
		BookTickerTree.ReplaceOrInsert(BookTickerItem{
			Symbol:      SymbolName(bookTicker.Symbol),
			BidPrice:    SymbolPrice(utils.ConvStrToFloat64(bookTicker.BidPrice)),
			BidQuantity: SymbolPrice(utils.ConvStrToFloat64(bookTicker.BidQuantity)),
			AskPrice:    SymbolPrice(utils.ConvStrToFloat64(bookTicker.AskPrice)),
			AskQuantity: SymbolPrice(utils.ConvStrToFloat64(bookTicker.AskQuantity)),
		})
	}
	return nil
}

func GetBookTickerTreeItem(symbol SymbolName) *BookTickerItem {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	item := BookTickerTree.Get(BookTickerItem{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*BookTickerItem)
}

// GetBookTickerTree returns the book ticker tree.
func GetBookTickerTree() *btree.BTree {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	return BookTickerTree
}

func SetBookTickerTree(tree *btree.BTree) {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	BookTickerTree = tree
}

func SetBookTickerTreeItem(item BookTickerItem) {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	BookTickerTree.ReplaceOrInsert(item)
}

func SearchBookTickerTreeBySymbol(symbol SymbolName) *btree.BTree {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	tree := btree.New(2)
	BookTickerTree.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

func SearchBookTickerTreeByBidPrice(symbol SymbolName, bidPrice SymbolPrice) *btree.BTree {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	tree := btree.New(2)
	BookTickerTree.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol && item.BidPrice == bidPrice {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

func SearchBookTickerTreeByAskPrice(symbol SymbolName, askPrice SymbolPrice) *btree.BTree {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	tree := btree.New(2)
	BookTickerTree.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol && item.AskPrice == askPrice {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

// ShowBookTickerTree prints the book ticker information for each item in the BookTickerTree.
func ShowBookTickerTree() {
	mu_bookticker_tree.Lock()
	defer mu_bookticker_tree.Unlock()
	BookTickerTree.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
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
