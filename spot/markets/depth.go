package markets

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
)

type (
	Price     float64
	DepthItem struct {
		Price           Price
		AskLastUpdateID int64
		AskQuantity     Price
		BidLastUpdateID int64
		BidQuantity     Price
	}
)

var (
	depths   = btree.New(2)
	mu_depth sync.Mutex
)

func (i *DepthItem) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItem).Price
}

func InitDepths(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_depth.Lock()
	defer mu_depth.Unlock()
	for _, bid := range res.Bids {
		depths.ReplaceOrInsert(&DepthItem{
			Price:           Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		depths.ReplaceOrInsert(&DepthItem{
			Price:           Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

func GetDepths() *btree.BTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	return depths
}

func SetDepths(tree *btree.BTree) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	depths = tree
}

func GetDepth(price Price) (*DepthItem, bool) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	item := depths.Get(&DepthItem{Price: price})
	if item == nil {
		return nil, false
	}
	return item.(*DepthItem), true
}

func SearchDepths(price Price) *btree.BTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	newTree := btree.New(2) // створюємо нове B-дерево

	depths.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItem)
		if item.Price == price {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func SetDepth(value DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	depths.ReplaceOrInsert(&DepthItem{
		Price:           value.Price,
		AskLastUpdateID: value.AskLastUpdateID,
		AskQuantity:     value.AskQuantity,
		BidLastUpdateID: value.BidLastUpdateID,
		BidQuantity:     value.BidQuantity,
	})
}

func GetDepthsByPrices(minPrice, maxPrice Price) *btree.BTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	newTree := btree.New(2) // створюємо нове B-дерево

	depths.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItem)
		if item.Price >= minPrice && item.Price <= maxPrice {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func GetDepthMaxBidQtyMaxAskQty() (*DepthItem, *DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	// Шукаємо вузол з максимальною ціною і ненульовим BidQuantity
	maxBidNode := &DepthItem{}
	maxAskNode := &DepthItem{}
	depths.Ascend(func(item btree.Item) bool {
		node := item.(*DepthItem)
		if node.BidQuantity != 0 && node.BidQuantity > maxBidNode.BidQuantity {
			maxBidNode = node
		}
		if node.AskQuantity != 0 && node.AskQuantity > maxAskNode.AskQuantity {
			maxAskNode = node
		}
		return true
	})
	return maxBidNode, maxAskNode
}

func GetMaxBidMinAsk() (*DepthItem, *DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	maxBid := &DepthItem{}
	minAsk := &DepthItem{}
	depths.Ascend(func(item btree.Item) bool {
		node := item.(*DepthItem)
		if node.BidQuantity != 0 && node.Price > maxBid.Price {
			maxBid = node
		}
		if minAsk.Price == 0 && node.AskQuantity != 0 {
			minAsk = node
		} else if node.AskQuantity != 0 && node.Price < minAsk.Price {
			minAsk = node
		}
		return true
	})
	return maxBid, minAsk
}

func ShowDepths() {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	depths.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItem)
		fmt.Println(
			"Price:", item.Price,
			"AskLastUpdateID:", item.AskLastUpdateID,
			"AskQuantity:", item.AskQuantity,
			"BidLastUpdateID:", item.BidLastUpdateID,
			"BidQuantity:", item.BidQuantity)
		return true
	})
}
