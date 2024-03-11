package info

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
)

var (
	depthTree = btree.New(2)
	mu_tree   sync.Mutex
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

func GetDepthTreeItem(price Price) (DepthRecord, bool) {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	item := depthTree.Get(DepthRecord{Price: price})
	if item == nil {
		return DepthRecord{}, false
	}
	return item.(DepthRecord), true
}

func SearchDepthTree(price Price) *btree.BTree {
	mu_map.Lock()
	defer mu_map.Unlock()
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

func SetDepthTreeItem(value DepthRecord) {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	depthTree.ReplaceOrInsert(DepthRecord{
		Price:           value.Price,
		AskLastUpdateID: value.AskLastUpdateID,
		AskQuantity:     value.AskQuantity,
		BidLastUpdateID: value.BidLastUpdateID,
		BidQuantity:     value.BidQuantity,
	})
}

func GetDepthTreeByPrices(minPrice, maxPrice Price) *btree.BTree {
	mu_map.Lock()
	defer mu_map.Unlock()
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

func ShowDepthTree() {
	mu_tree.Lock()
	defer mu_tree.Unlock()
	depthTree.Ascend(func(i btree.Item) bool {
		item := i.(DepthRecord)
		fmt.Println(
			"Price:", item.Price,
			"AskLastUpdateID:", item.AskLastUpdateID,
			"AskQuantity:", item.AskQuantity,
			"BidLastUpdateID:", item.BidLastUpdateID,
			"BidQuantity:", item.BidQuantity)
		return true
	})
}
