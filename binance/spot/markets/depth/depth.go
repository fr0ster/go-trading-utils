package depth

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	"github.com/adshao/go-binance/v2/futures"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	// DepthBTree depth_interface.DepthBTree
	DepthBTree struct {
		client *binance.Client
		btree.BTree
		mutex  sync.Mutex
		degree int
		depth_interface.AskLastUpdateID
		depth_interface.BidLastUpdateID
	}
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree int) *DepthBTree {
	return &DepthBTree{
		client: nil,
		BTree:  *btree.New(int(degree)),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

// Init implements depth_interface.Depths.
func (d *DepthBTree) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) error {
	futures.UseTestnet = UseTestnet
	d.client = binance.NewClient(apt_key, secret_key)
	res, err :=
		d.client.NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, bid := range res.Bids {
		d.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:       types.Price(utils.ConvStrToFloat64(bid.Price)),
			BidQuantity: types.Price(utils.ConvStrToFloat64(bid.Quantity)),
		})
	}
	for _, ask := range res.Asks {
		d.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:       types.Price(utils.ConvStrToFloat64(ask.Price)),
			AskQuantity: types.Price(utils.ConvStrToFloat64(ask.Quantity)),
		})
	}
	return nil
}

// DeleteItem implements depth_interface.Depths.
func (d *DepthBTree) DeleteItem(value *depth_interface.DepthItemType) bool {
	item := d.BTree.Delete(value)
	return item != nil
}

// GetAskQtyLocalMaxima implements depth_interface.Depths.
func (d *DepthBTree) GetAskQtyLocalMaxima() *btree.BTree {
	maximaTree := New(int(d.degree))
	var prev, current, next *depth_interface.DepthItemType
	d.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
		if current != nil && prev != nil && current.AskQuantity > prev.AskQuantity && current.AskQuantity > next.AskQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return &maximaTree.BTree
}

// GetBidQtyLocalMaxima implements depth_interface.Depths.
func (d *DepthBTree) GetBidQtyLocalMaxima() *btree.BTree {
	maximaTree := New(int(d.degree))
	var prev, current, next *depth_interface.DepthItemType
	d.Ascend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
		if current != nil && prev != nil && current.BidQuantity > prev.BidQuantity && current.BidQuantity > next.BidQuantity {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return &maximaTree.BTree
}

// GetItem implements depth_interface.Depths.
func (d *DepthBTree) GetItem(price types.Price) *depth_interface.DepthItemType {
	item := d.BTree.Get(&depth_interface.DepthItemType{Price: price})
	return item.(*depth_interface.DepthItemType)
}

// GetMaxAsks implements depth_interface.Depths.
func (d *DepthBTree) GetMaxAsks() *depth_interface.DepthItemType {
	ask := depth_interface.DepthItemType{}
	d.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		if item.AskQuantity != 0 {
			ask = *item
		}
		return true
	})
	return &ask
}

// GetMaxBidMinAsk implements depth_interface.Depths.
func (d *DepthBTree) GetMaxBidMinAsk() (maxBid *depth_interface.DepthItemType, minAsk *depth_interface.DepthItemType) {
	maxBid = &depth_interface.DepthItemType{}
	minAsk = &depth_interface.DepthItemType{}
	d.Ascend(func(item btree.Item) bool {
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

// GetMaxBidQtyMaxAskQty implements depth_interface.Depths.
func (d *DepthBTree) GetMaxBidQtyMaxAskQty() (maxBidNode *depth_interface.DepthItemType, maxAskNode *depth_interface.DepthItemType) {
	maxBidNode = &depth_interface.DepthItemType{}
	maxAskNode = &depth_interface.DepthItemType{}
	d.Ascend(func(item btree.Item) bool {
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

// GetMaxBids implements depth_interface.Depths.
func (d *DepthBTree) GetMaxBids() *depth_interface.DepthItemType {
	bid := &depth_interface.DepthItemType{}
	d.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		if item.BidQuantity != 0 {
			bid = item
		}
		return true
	})
	return bid
}

// Lock implements depth_interface.Depths.
// Subtle: this method shadows the method (Mutex).Lock of DepthBTree.Mutex.
func (d *DepthBTree) Lock() {
	d.mutex.Lock()
}

// SetItem implements depth_interface.Depths.
func (d *DepthBTree) SetItem(value depth_interface.DepthItemType) {
	d.BTree.ReplaceOrInsert(&depth_interface.DepthItemType{
		Price:       value.Price,
		AskQuantity: value.AskQuantity,
		BidQuantity: value.BidQuantity,
	})
}

// Show implements depth_interface.Depths.
func (d *DepthBTree) Show() {
	d.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		fmt.Println(
			"Price:", item.Price,
			"AskQuantity:", item.AskQuantity,
			"BidQuantity:", item.BidQuantity)
		return true
	})
}

// Unlock implements depth_interface.Depths.
// Subtle: this method shadows the method (Mutex).Unlock of DepthBTree.Mutex.
func (d *DepthBTree) Unlock() {
	d.mutex.Unlock()
}

// UpdateAsk implements depth_interface.Depths.
func (d *DepthBTree) UpdateAsk(ask common.PriceLevel, askLastUpdateID depth_interface.AskLastUpdateID) (err error) {
	price, quantity, err := ask.Parse()
	if err != nil {
		return
	}
	value := d.GetItem(types.Price(price))
	d.AskLastUpdateID = askLastUpdateID
	if value != nil {
		value.AskQuantity += types.Price(quantity)
	} else {
		value =
			&depth_interface.DepthItemType{
				Price:       types.Price(price),
				AskQuantity: types.Price(quantity),
				BidQuantity: 0,
			}
	}
	d.SetItem(*value)
	return nil
}

// UpdateBid implements depth_interface.Depths.
func (d *DepthBTree) UpdateBid(bid common.PriceLevel, bidLastUpdateID depth_interface.BidLastUpdateID) (err error) {
	price, quantity, err := bid.Parse()
	if err != nil {
		return
	}
	value := d.GetItem(types.Price(price))
	d.BidLastUpdateID = bidLastUpdateID
	if value != nil {
		value.BidQuantity += types.Price(quantity)
	} else {
		value =
			&depth_interface.DepthItemType{
				Price:       types.Price(price),
				AskQuantity: types.Price(quantity),
				BidQuantity: 0,
			}
	}
	d.SetItem(*value)
	return nil
}
