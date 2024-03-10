package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
	"github.com/google/btree"
)

var (
	bookTickerTree = btree.New(2)
	mu_tree        sync.Mutex
)

func (i DepthRecord) Less(than btree.Item) bool {
	return i.Price < than.(DepthRecord).Price
}

func InitDepthTree(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_tree.Lock()
	defer mu_tree.Unlock()
	for _, bid := range res.Bids {
		bookTickerTree.ReplaceOrInsert(DepthRecord{
			Price:           Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		bookTickerTree.ReplaceOrInsert(DepthRecord{
			Price:           Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

func GetBookTickerTree() *btree.BTree {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	return bookTickerTree
}

func SearchBookTickerTree(price Price) *DepthRecord {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	return bookTickerTree.Get(DepthRecord{Price: price}).(*DepthRecord)
}

func SearchBookTickerTreeByPrices(low Price, high Price) *DepthRecord {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	var result *DepthRecord
	bookTickerTree.AscendGreaterOrEqual(DepthRecord{Price: low}, func(i btree.Item) bool {
		if i.(DepthRecord).Price > high {
			return false
		}
		result = i.(*DepthRecord)
		return true
	})
	return result
}
