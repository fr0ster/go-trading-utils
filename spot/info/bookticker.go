package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
)

type (
	BookTickerItem struct {
		Symbol      SymbolType
		BidPrice    PriceType
		BidQuantity PriceType
		AskPrice    PriceType
		AskQuantity PriceType
	}
	PriceType  float64
	SymbolType string
)

var (
	bookTickers   = btree.New(2) // Book ticker tree
	mu_bookticker sync.Mutex     // Mutex for book ticker tree
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b BookTickerItem) Less(than btree.Item) bool {
	return b.Symbol < than.(BookTickerItem).Symbol
}

// InitBookTicker initializes the book ticker tree with prices.
// It retrieves the book tickers for the given symbol from the Binance client
// and inserts them into the book ticker tree.
func InitBookTicker(client *binance.Client, symbolname string) (err error) {
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	for _, bookTicker := range bookTickerList {
		bookTickers.ReplaceOrInsert(BookTickerItem{
			Symbol:      SymbolType(bookTicker.Symbol),
			BidPrice:    PriceType(utils.ConvStrToFloat64(bookTicker.BidPrice)),
			BidQuantity: PriceType(utils.ConvStrToFloat64(bookTicker.BidQuantity)),
			AskPrice:    PriceType(utils.ConvStrToFloat64(bookTicker.AskPrice)),
			AskQuantity: PriceType(utils.ConvStrToFloat64(bookTicker.AskQuantity)),
		})
	}
	return nil
}

func GetBookTicker(symbol SymbolType) *BookTickerItem {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	item := bookTickers.Get(BookTickerItem{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*BookTickerItem)
}

// GetBookTickers returns the book ticker tree.
func GetBookTickers() *btree.BTree {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	return bookTickers
}

func SetBookTickers(tree *btree.BTree) {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	bookTickers = tree
}

func SetBookTicker(item BookTickerItem) {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	bookTickers.ReplaceOrInsert(item)
}

func SearchBookTickersBySymbol(symbol SymbolType) *btree.BTree {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	tree := btree.New(2)
	bookTickers.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

func SearchBookTickersByBidPrice(symbol SymbolType, bidPrice PriceType) *btree.BTree {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	tree := btree.New(2)
	bookTickers.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol && item.BidPrice == bidPrice {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

func SearchBookTickersByAskPrice(symbol SymbolType, askPrice PriceType) *btree.BTree {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	tree := btree.New(2)
	bookTickers.Ascend(func(i btree.Item) bool {
		item := i.(BookTickerItem)
		if item.Symbol == symbol && item.AskPrice == askPrice {
			tree.ReplaceOrInsert(item)
		}
		return true
	})
	return tree
}

// ShowBookTickers prints the book ticker information for each item in the BookTickerTree.
func ShowBookTickers() {
	mu_bookticker.Lock()
	defer mu_bookticker.Unlock()
	bookTickers.Ascend(func(i btree.Item) bool {
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
