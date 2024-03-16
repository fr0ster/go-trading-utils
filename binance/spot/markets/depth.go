package markets

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	Price      float64
	DepthBTree struct {
		*btree.BTree
		sync.Mutex
	}
	DepthItemType struct {
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
		Mutex: sync.Mutex{},
	}
}

func (i *DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItemType).Price
}

func (i *DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItemType).Price
}

func (tree *DepthBTree) Lock() {
	tree.Mutex.Lock()
}

func (tree *DepthBTree) Unlock() {
	tree.Mutex.Unlock()
}

func (tree *DepthBTree) Init(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	for _, bid := range res.Bids {
		tree.ReplaceOrInsert(&DepthItemType{
			Price:           Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		tree.ReplaceOrInsert(&DepthItemType{
			Price:           Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

func (tree *DepthBTree) GetItem(price Price) (*DepthItemType, bool) {
	item := tree.Get(&DepthItemType{Price: price})
	if item == nil {
		return nil, false
	}
	return item.(*DepthItemType), true
}

func (tree *DepthBTree) SetItem(value DepthItemType) {
	tree.ReplaceOrInsert(&DepthItemType{
		Price:           value.Price,
		AskLastUpdateID: value.AskLastUpdateID,
		AskQuantity:     value.AskQuantity,
		BidLastUpdateID: value.BidLastUpdateID,
		BidQuantity:     value.BidQuantity,
	})
}

func (tree *DepthBTree) GetByPrices(minPrice, maxPrice Price) *DepthBTree {
	newTree := DepthNew(2) // створюємо нове B-дерево

	tree.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItemType)
		if item.Price >= minPrice && item.Price <= maxPrice {
			newTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func (tree *DepthBTree) GetMaxBids() *DepthBTree {
	bids := DepthNew(2)
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItemType)
		if item.BidQuantity != 0 {
			bids.ReplaceOrInsert(item)
		}
		return true
	})
	return bids
}

func (tree *DepthBTree) GetMaxAsks() *DepthBTree {
	asks := DepthNew(2)
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItemType)
		if item.AskQuantity != 0 {
			asks.ReplaceOrInsert(item)
		}
		return true
	})
	return asks
}

func (tree *DepthBTree) GetMaxBidQtyMaxAskQty() (maxBidNode *DepthItemType, maxAskNode *DepthItemType) {
	// Шукаємо вузол з максимальною ціною і ненульовим BidQuantity
	maxBidNode = &DepthItemType{}
	maxAskNode = &DepthItemType{}
	tree.Ascend(func(item btree.Item) bool {
		node := item.(*DepthItemType)
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

func (tree *DepthBTree) GetMaxBidMinAsk() (maxBid *DepthItemType, minAsk *DepthItemType) {
	maxBid = &DepthItemType{}
	minAsk = &DepthItemType{}
	tree.Ascend(func(item btree.Item) bool {
		node := item.(*DepthItemType)
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

func (tree *DepthBTree) GetBidQtyLocalMaxima() *DepthBTree {
	maximaTree := DepthNew(2)
	var prev, current, next *DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if current != nil && prev != nil && current.BidQuantity > prev.BidQuantity && current.BidQuantity > next.BidQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

func (tree *DepthBTree) GetAskQtyLocalMaxima() *DepthBTree {
	maximaTree := DepthNew(2)
	var prev, current, next *DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if current != nil && prev != nil && current.AskQuantity > prev.AskQuantity && current.AskQuantity > next.AskQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

func (tree *DepthBTree) GetBidQtyLocalMinima() *DepthBTree {
	minimaTree := DepthNew(2)
	var prev, current, next *DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if current != nil && prev != nil && current.BidQuantity < prev.BidQuantity && current.BidQuantity < next.BidQuantity {
			minimaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return minimaTree
}

func (tree *DepthBTree) GetAskQtyLocalMinima() *DepthBTree {
	minimaTree := DepthNew(2)
	var prev, current, next *DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*DepthItemType)
		if current != nil && prev != nil && current.AskQuantity < prev.AskQuantity && current.AskQuantity < next.AskQuantity {
			minimaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return minimaTree
}

func (tree *DepthBTree) Show() {
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*DepthItemType)
		fmt.Println(
			"Price:", item.Price,
			"AskLastUpdateID:", item.AskLastUpdateID,
			"AskQuantity:", item.AskQuantity,
			"BidLastUpdateID:", item.BidLastUpdateID,
			"BidQuantity:", item.BidQuantity)
		return true
	})
}
