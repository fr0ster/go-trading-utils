package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
	"github.com/google/btree"
)

var (
	depthTree = btree.New(2)
	mu_tree   *sync.Mutex
)

func (i DepthRecord) Less(than btree.Item) bool {
	return i.Price < than.(DepthRecord).Price
}

func InitDepthTree(client *binance.Client, mu *sync.Mutex, symbolname string) (err error) {
	mu_tree = mu
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
		depthTree.ReplaceOrInsert(DepthRecord{
			Price:           Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		depthTree.ReplaceOrInsert(DepthRecord{
			Price:           Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

func GetDepthTree() *btree.BTree {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	return depthTree
}

func SetDepthTree(tree *btree.BTree) {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	depthTree = tree
}

func SearchDepthTree(price Price) *btree.BTree {
	newTree := btree.New(2) // створюємо нове B-дерево

	depthTree.Ascend(func(i btree.Item) bool {
		item := i.(DepthRecord)
		if item.Price == price {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func SearchDepthTreeByPrices(minPrice, maxPrice Price) *btree.BTree {
	newTree := btree.New(2) // створюємо нове B-дерево

	depthTree.Ascend(func(i btree.Item) bool {
		item := i.(DepthRecord)
		if item.Price >= minPrice && item.Price <= maxPrice {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}
