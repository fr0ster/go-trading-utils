package depth

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	DepthBTree depth_interface.DepthBTree
)

// DepthBTree - B-дерево для зберігання стакана заявок
func DepthNew(degree int) *DepthBTree {
	return &DepthBTree{
		BTree:  *btree.New(int(degree)),
		Mutex:  sync.Mutex{},
		Degree: depth_interface.Degree(degree),
	}
}

func (tree *DepthBTree) Lock() {
	tree.Mutex.Lock()
}

func (tree *DepthBTree) Unlock() {
	tree.Mutex.Unlock()
}

func (tree *DepthBTree) Init(apt_key, secret_key, symbolname string, UseTestnet bool) *DepthBTree {
	binance.UseTestnet = UseTestnet
	res, err :=
		binance.NewClient(apt_key, secret_key).NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return nil
	}
	for _, bid := range res.Bids {
		tree.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:           types.Price(utils.ConvStrToFloat64(bid.Price)),
			BidLastUpdateID: res.LastUpdateID,
			BidQuantity:     types.Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		tree.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:           types.Price(utils.ConvStrToFloat64(ask.Price)),
			AskLastUpdateID: res.LastUpdateID,
			AskQuantity:     types.Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

func (tree *DepthBTree) GetItem(price types.Price) (*depth_interface.DepthItemType, bool) {
	item := tree.BTree.Get(&depth_interface.DepthItemType{Price: price})
	if item == nil {
		return nil, false
	}
	return item.(*depth_interface.DepthItemType), true
}

func (tree *DepthBTree) SetItem(value depth_interface.DepthItemType) {
	tree.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
		Price:           value.Price,
		AskLastUpdateID: value.AskLastUpdateID,
		AskQuantity:     value.AskQuantity,
		BidLastUpdateID: value.BidLastUpdateID,
		BidQuantity:     value.BidQuantity,
	})
}

func (tree *DepthBTree) GetByPrices(minPrice, maxPrice types.Price) *DepthBTree {
	newTree := DepthNew(int(tree.Degree)) // створюємо нове B-дерево

	tree.BTree.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		if item.Price >= minPrice && item.Price <= maxPrice {
			newTree.BTree.ReplaceOrInsert(item) // додаємо вузол до нового дерева, якщо він відповідає умовам
		}
		return true
	})

	return newTree
}

func (tree *DepthBTree) GetMaxBids() *DepthBTree {
	bids := DepthNew(int(tree.Degree))
	tree.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
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
		item := i.(*depth_interface.DepthItemType)
		if item.AskQuantity != 0 {
			asks.ReplaceOrInsert(item)
		}
		return true
	})
	return asks
}

func (tree *DepthBTree) GetMaxBidQtyMaxAskQty() (maxBidNode *depth_interface.DepthItemType, maxAskNode *depth_interface.DepthItemType) {
	// Шукаємо вузол з максимальною ціною і ненульовим BidQuantity
	maxBidNode = &depth_interface.DepthItemType{}
	maxAskNode = &depth_interface.DepthItemType{}
	tree.Ascend(func(item btree.Item) bool {
		node := item.(*depth_interface.DepthItemType)
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

func (tree *DepthBTree) GetMaxBidMinAsk() (maxBid *depth_interface.DepthItemType, minAsk *depth_interface.DepthItemType) {
	maxBid = &depth_interface.DepthItemType{}
	minAsk = &depth_interface.DepthItemType{}
	tree.Ascend(func(item btree.Item) bool {
		node := item.(*depth_interface.DepthItemType)
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
	var prev, current, next *depth_interface.DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
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
	var prev, current, next *depth_interface.DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
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
	var prev, current, next *depth_interface.DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
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
	var prev, current, next *depth_interface.DepthItemType
	tree.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
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
		item := i.(*depth_interface.DepthItemType)
		fmt.Println(
			"Price:", item.Price,
			"AskLastUpdateID:", item.AskLastUpdateID,
			"AskQuantity:", item.AskQuantity,
			"BidLastUpdateID:", item.BidLastUpdateID,
			"BidQuantity:", item.BidQuantity)
		return true
	})
}
