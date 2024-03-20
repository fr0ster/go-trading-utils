package depth

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	Depth struct {
		symbol          string
		asks            btree.BTree
		bids            btree.BTree
		mutex           sync.Mutex
		degree          int
		round           int
		limit           int
		AskLastUpdateID int64
		BidLastUpdateID int64
	}
	// DepthBTree btree.BTree
)

// DepthItemType - тип для зберігання заявок в стакані
func (i *Depth) Less(than btree.Item) bool {
	return i.symbol < than.(*Depth).symbol
}

func (i *Depth) Equal(than btree.Item) bool {
	return i.symbol == than.(*Depth).symbol
}

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree, round, limit int, symbol string) *Depth {
	return &Depth{
		symbol: symbol,
		asks:   *btree.New(degree),
		bids:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
		round:  round,
		limit:  limit,
	}
}

// GetAsks implements depth_interface.Depths.
func (d *Depth) GetAsks() *btree.BTree {
	return &d.asks
}

// GetBids implements depth_interface.Depths.
func (d *Depth) GetBids() *btree.BTree {
	return &d.bids
}

// SetAsks implements depth_interface.Depths.
func (d *Depth) SetAsks(asks *btree.BTree) {
	d.asks = *asks
}

// SetBids implements depth_interface.Depths.
func (d *Depth) SetBids(bids *btree.BTree) {
	d.bids = *bids
}

func (d *Depth) Init(apt_key, secret_key, symbolname string, UseTestnet bool) (err error) {
	futures.UseTestnet = UseTestnet
	client := futures.NewClient(apt_key, secret_key)
	res, err :=
		client.NewDepthService().
			Symbol(string(symbolname)).
			Limit(d.limit).
			Do(context.Background())
	if err != nil {
		return err
	}
	for _, bid := range res.Bids {
		d.bids.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:    utils.ConvStrToFloat64(bid.Price),
			Quantity: utils.ConvStrToFloat64(bid.Quantity),
		})
	}
	for _, ask := range res.Asks {
		d.asks.ReplaceOrInsert(&depth_interface.DepthItemType{
			Price:    utils.ConvStrToFloat64(ask.Price),
			Quantity: utils.ConvStrToFloat64(ask.Quantity),
		})
	}
	return nil
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
func (d *Depth) GetAsk(price float64) *depth_interface.DepthItemType {
	item := d.asks.Get(&depth_interface.DepthItemType{Price: price})
	if item == nil {
		return nil
	}
	return item.(*depth_interface.DepthItemType)
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price float64) *depth_interface.DepthItemType {
	item := d.bids.Get(&depth_interface.DepthItemType{Price: price})
	if item == nil {
		return nil
	}
	return item.(*depth_interface.DepthItemType)
}

// SetAsk implements depth_interface.Depths.
func (d *Depth) SetAsk(value depth_interface.DepthItemType) {
	d.asks.ReplaceOrInsert(&value)
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(value depth_interface.DepthItemType) {
	d.bids.ReplaceOrInsert(&value)
}

// UpdateAsk implements depth_interface.Depths.
func (d *Depth) UpdateAsk(price float64, quantity float64) {
	old := d.asks.Get(&depth_interface.DepthItemType{Price: price})
	if old != nil {
		old := old.(*depth_interface.DepthItemType)
		d.asks.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity + old.Quantity})
	} else {
		d.asks.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity})
	}
}

// UpdateBid implements depth_interface.Depths.
func (d *Depth) UpdateBid(price float64, quantity float64) {

	old := d.bids.Get(&depth_interface.DepthItemType{Price: price})
	if old != nil {
		old := old.(*depth_interface.DepthItemType)
		d.bids.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity + old.Quantity})
	} else {
		d.bids.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity})
	}
}

// GetMaxAsks implements depth_interface.Depths.
func (d *Depth) GetMaxAsks() *depth_interface.DepthItemType {
	item := d.asks.Max()
	return item.(*depth_interface.DepthItemType)
}

// GetMaxBids implements depth_interface.Depths.
func (d *Depth) GetMaxBids() *depth_interface.DepthItemType {
	item := d.bids.Max()
	if item == nil {
		return nil
	}
	return item.(*depth_interface.DepthItemType)
}

// GetMinAsks implements depth_interface.Depths.
func (d *Depth) GetMinAsks() *depth_interface.DepthItemType {
	item := d.asks.Min()
	return item.(*depth_interface.DepthItemType)
}

// GetMinBids implements depth_interface.Depths.
func (d *Depth) GetMinBids() *depth_interface.DepthItemType {
	item := d.bids.Min()
	if item == nil {
		return nil
	}
	return item.(*depth_interface.DepthItemType)
}

// GetBidLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetBidLocalMaxima() *btree.BTree {
	maximaTree := btree.New(d.degree)
	var prev, current, next *depth_interface.DepthItemType
	d.BidAscend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
		if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
			(current != nil && prev == nil && current.Quantity > next.Quantity) {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

// GetAskLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetAskLocalMaxima() *btree.BTree {
	maximaTree := btree.New(d.degree)
	var prev, current, next *depth_interface.DepthItemType
	d.AskAscend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
		if (current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity) ||
			(current != nil && prev == nil && current.Quantity > next.Quantity) {
			maximaTree.ReplaceOrInsert(current)
		}
		prev = current
		current = next
		return true
	})
	return maximaTree
}

// Lock implements depth_interface.Depths.
func (d *Depth) Lock() {
	d.mutex.Lock()
}

// Unlock implements depth_interface.Depths.
func (d *Depth) Unlock() {
	d.mutex.Unlock()
}
