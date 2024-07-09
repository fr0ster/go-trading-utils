package depth

import (
	"errors"
	"sync"

	pair_price_types "github.com/fr0ster/go-trading-utils/types/pair_price"

	"github.com/google/btree"
)

type (
	// DepthItemType - тип для зберігання заявок в стакані
	Depth struct {
		symbol            string
		asks              *btree.BTree
		asksSummaQuantity float64
		bids              *btree.BTree
		bidsSummaQuantity float64
		mutex             *sync.Mutex
		LastUpdateID      int64
	}
)

func (i *Depth) Less(than btree.Item) bool {
	return i.symbol < than.(*Depth).symbol
}

func (i *Depth) Equal(than btree.Item) bool {
	return i.symbol == than.(*Depth).symbol
}

// GetAsks implements depth_interface.Depths.
func (d *Depth) GetAsks() *btree.BTree {
	return d.asks
}

// GetBids implements depth_interface.Depths.
func (d *Depth) GetBids() *btree.BTree {
	return d.bids
}

// SetAsks implements depth_interface.Depths.
func (d *Depth) SetAsks(asks *btree.BTree) {
	d.asks = asks
}

// SetBids implements depth_interface.Depths.
func (d *Depth) SetBids(bids *btree.BTree) {
	d.bids = bids
}

// DeleteAsk implements depth_interface.Depths.
func (d *Depth) DeleteAsk(price float64) {
	d.asks.Delete(&pair_price_types.PairPrice{Price: price})
}

// DeleteBid implements depth_interface.Depths.
func (d *Depth) DeleteBid(price float64) {
	d.bids.Delete(&pair_price_types.PairPrice{Price: price})
}

// ClearAsks implements depth_interface.Depths.
func (d *Depth) ClearAsks() {
	d.asks.Clear(false)
}

// ClearBids implements depth_interface.Depths.
func (d *Depth) ClearBids() {
	d.bids.Clear(false)
}

// AskAscend implements depth_interface.Depths.
func (d *Depth) AskAscend(iter func(btree.Item) bool) {
	d.asks.Ascend(iter)
}

// AskDescend implements depth_interface.Depths.
func (d *Depth) AskDescend(iter func(btree.Item) bool) {
	d.asks.Descend(iter)
}

// BidAscend implements depth_interface.Depths.
func (d *Depth) BidAscend(iter func(btree.Item) bool) {
	d.bids.Ascend(iter)
}

// BidDescend implements depth_interface.Depths.
func (d *Depth) BidDescend(iter func(btree.Item) bool) {
	d.bids.Descend(iter)
}

// GetAsk implements depth_interface.Depths.
func (d *Depth) GetAsk(price float64) btree.Item {
	item := d.asks.Get(&pair_price_types.PairPrice{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price float64) btree.Item {
	item := d.bids.Get(&pair_price_types.PairPrice{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// SetAsk implements depth_interface.Depths.
func (d *Depth) SetAsk(price float64, quantity float64) {
	d.asks.ReplaceOrInsert(&pair_price_types.PairPrice{Price: price, Quantity: quantity})
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(price float64, quantity float64) {
	d.bids.ReplaceOrInsert(&pair_price_types.PairPrice{Price: price, Quantity: quantity})
}

func (d *Depth) AsksSummaQuantity() float64 {
	return d.asksSummaQuantity
}

func (d *Depth) BidsSummaQuantity() float64 {
	return d.bidsSummaQuantity
}

// RestrictAsk implements depth_interface.Depths.
func (d *Depth) RestrictAsk(price float64) {
	d.asks.Ascend(func(i btree.Item) bool {
		if i.(*pair_price_types.PairPrice).Price < price {
			d.asks.Delete(i)
			return false
		}
		return true
	})
}

// RestrictBid implements depth_interface.Depths.
func (d *Depth) RestrictBid(price float64) {
	d.bids.Ascend(func(i btree.Item) bool {
		if i.(*pair_price_types.PairPrice).Price > price {
			d.bids.Delete(i)
			return false
		}
		return true
	})
}

// UpdateAsk implements depth_interface.Depths.
func (d *Depth) UpdateAsk(price float64, quantity float64) bool {
	old := d.asks.Get(&pair_price_types.PairPrice{Price: price})
	if old != nil && old.(*pair_price_types.PairPrice).Quantity == quantity {
		d.asksSummaQuantity -= old.(*pair_price_types.PairPrice).Quantity
		return false
	}
	if quantity == 0 {
		d.asks.Delete(&pair_price_types.PairPrice{Price: price})
	} else {
		d.asksSummaQuantity += quantity
		d.asks.ReplaceOrInsert(&pair_price_types.PairPrice{Price: price, Quantity: quantity})
	}
	return true
}

// UpdateBid implements depth_interface.Depths.
func (d *Depth) UpdateBid(price float64, quantity float64) bool {
	old := d.bids.Get(&pair_price_types.PairPrice{Price: price})
	if old != nil && old.(*pair_price_types.PairPrice).Quantity == quantity {
		d.bidsSummaQuantity -= old.(*pair_price_types.PairPrice).Quantity
		return false
	}
	if quantity == 0 {
		d.bids.Delete(&pair_price_types.PairPrice{Price: price})
	} else {
		d.bidsSummaQuantity += quantity
		d.bids.ReplaceOrInsert(&pair_price_types.PairPrice{Price: price, Quantity: quantity})
	}
	return true
}

// Lock implements depth_interface.Depths.
func (d *Depth) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Depth) Unlock() {
	d.mutex.Unlock()
}

// Symbol implements depth_interface.Depths.
func (d *Depth) Symbol() string {
	return d.symbol
}

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree int, symbol string) *Depth {
	return &Depth{
		symbol: symbol,
		asks:   btree.New(degree),
		bids:   btree.New(degree),
		mutex:  &sync.Mutex{},
	}
}

func Binance2BookTicker(binanceDepth interface{}) (*pair_price_types.PairPrice, error) {
	switch binanceDepth := binanceDepth.(type) {
	case *pair_price_types.PairPrice:
		return binanceDepth, nil
	}
	return nil, errors.New("it's not a DepthItemType")
}
