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
	Price      float64
	DepthBTree struct {
		*btree.BTree
	}
	DepthItem struct {
		Price           Price
		AskLastUpdateID int64
		AskQuantity     Price
		BidLastUpdateID int64
		BidQuantity     Price
	}
)

func DepthNew(degree int) *DepthBTree {
	return &DepthBTree{
		BTree: btree.New(degree),
	}
}

var mu_depth sync.Mutex

func (i *DepthItem) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItem).Price
}

func (tree *DepthBTree) InitDepths(client *binance.Client, symbolname string) (err error) {
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
		tree.ReplaceOrInsert(&DepthItem{
			Price:           Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		tree.ReplaceOrInsert(&DepthItem{
			Price:           Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

// func GetDepths() *DepthBTree {
// 	mu_depth.Lock()
// 	defer mu_depth.Unlock()
// 	return tree
// }

// func SetDepths(tree *DepthBTree) {
// 	mu_depth.Lock()
// 	defer mu_depth.Unlock()
// 	tree = tree
// }

func (tree *DepthBTree) GetDepth(price Price) (*DepthItem, bool) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	item := tree.Get(&DepthItem{Price: price})
	if item == nil {
		return nil, false
	}
	return item.(*DepthItem), true
}

func (tree *DepthBTree) SetDepth(value DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	tree.ReplaceOrInsert(&DepthItem{
		Price:           value.Price,
		AskLastUpdateID: value.AskLastUpdateID,
		AskQuantity:     value.AskQuantity,
		BidLastUpdateID: value.BidLastUpdateID,
		BidQuantity:     value.BidQuantity,
	})
}

func (tree *DepthBTree) GetDepthsByPrices(minPrice, maxPrice Price) *DepthBTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	newTree := DepthNew(2) // створюємо нове B-дерево

	tree.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItem)
		if item.Price >= minPrice && item.Price <= maxPrice {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func (tree *DepthBTree) GetDepthMaxBidQtyMaxAskQty() (maxBidNode *DepthItem, maxAskNode *DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	// Шукаємо вузол з максимальною ціною і ненульовим BidQuantity
	maxBidNode = &DepthItem{}
	maxAskNode = &DepthItem{}
	tree.Ascend(func(item btree.Item) bool {
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

func (tree *DepthBTree) GetDepthMaxBidMinAsk() (maxBid *DepthItem, minAsk *DepthItem) {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	maxBid = &DepthItem{}
	minAsk = &DepthItem{}
	tree.Ascend(func(item btree.Item) bool {
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

func (tree *DepthBTree) GetDepthBidQtyLocalMaxima() *DepthBTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	maximaTree := DepthNew(2)
	var prev, current, next *DepthItem
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItem)
		if current != nil && prev != nil && current.BidQuantity > prev.BidQuantity && current.BidQuantity > next.BidQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

func (tree *DepthBTree) GetDepthAskQtyLocalMaxima() *DepthBTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	maximaTree := DepthNew(2)
	var prev, current, next *DepthItem
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItem)
		if current != nil && prev != nil && current.AskQuantity > prev.AskQuantity && current.AskQuantity > next.AskQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

func (tree *DepthBTree) GetDepthBidQtyLocalMinima() *DepthBTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	minimaTree := DepthNew(2)
	var prev, current, next *DepthItem
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItem)
		if current != nil && prev != nil && current.BidQuantity < prev.BidQuantity && current.BidQuantity < next.BidQuantity {
			minimaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return minimaTree
}

func (tree *DepthBTree) GetDepthAskQtyLocalMinima() *DepthBTree {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	minimaTree := DepthNew(2)
	var prev, current, next *DepthItem
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItem)
		if current != nil && prev != nil && current.AskQuantity < prev.AskQuantity && current.AskQuantity < next.AskQuantity {
			minimaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return minimaTree
}

func (tree *DepthBTree) ShowDepths() {
	mu_depth.Lock()
	defer mu_depth.Unlock()
	tree.Ascend(func(i btree.Item) bool {
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
