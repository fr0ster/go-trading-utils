package depth

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	Depth struct {
		client          *binance.Client
		asks            btree.BTree
		bids            btree.BTree
		mutex           sync.Mutex
		degree          int
		round           int
		AskLastUpdateID int64
		BidLastUpdateID int64
	}
	// DepthBTree btree.BTree
)

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree, round int) *Depth {
	return &Depth{
		client: nil,
		asks:   *btree.New(degree),
		bids:   *btree.New(degree),
		mutex:  sync.Mutex{},
		degree: degree,
		round:  round,
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
	binance.UseTestnet = UseTestnet
	d.client = binance.NewClient(apt_key, secret_key)
	res, err :=
		d.client.NewDepthService().
			Symbol(string(symbolname)).
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
	return d.asks.Get(&depth_interface.DepthItemType{Price: price}).(*depth_interface.DepthItemType)
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price float64) *depth_interface.DepthItemType {
	return d.bids.Get(&depth_interface.DepthItemType{Price: price}).(*depth_interface.DepthItemType)
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
	d.asks.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity})
}

// UpdateBid implements depth_interface.Depths.
func (d *Depth) UpdateBid(price float64, quantity float64) {
	d.bids.ReplaceOrInsert(&depth_interface.DepthItemType{Price: price, Quantity: quantity})
}

// GetMaxAsks implements depth_interface.Depths.
func (d *Depth) GetMaxAsks() *depth_interface.DepthItemType {
	return d.asks.Max().(*depth_interface.DepthItemType)
}

// GetMaxBids implements depth_interface.Depths.
func (d *Depth) GetMaxBids() *depth_interface.DepthItemType {
	return d.bids.Max().(*depth_interface.DepthItemType)
}

// GetMinAsks implements depth_interface.Depths.
func (d *Depth) GetMinAsks() *depth_interface.DepthItemType {
	return d.asks.Min().(*depth_interface.DepthItemType)
}

// GetMinBids implements depth_interface.Depths.
func (d *Depth) GetMinBids() *depth_interface.DepthItemType {
	return d.bids.Min().(*depth_interface.DepthItemType)
}

// GetBidLocalMaxima implements depth_interface.Depths.
func (d *Depth) GetBidLocalMaxima() *btree.BTree {
	maximaTree := btree.New(d.degree)
	var prev, current, next *depth_interface.DepthItemType
	d.BidAscend(func(a btree.Item) bool {
		next = a.(*depth_interface.DepthItemType)
		if current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity {
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
		if current != nil && prev != nil && current.Quantity > prev.Quantity && current.Quantity > next.Quantity {
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
