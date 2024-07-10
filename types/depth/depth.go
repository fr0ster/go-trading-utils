package depth

import (
	"errors"
	"sync"

	"github.com/google/btree"
)

type (
	QuantityItem struct {
		Quantity float64
		Price    float64
	}
	DepthItem struct {
		Price    float64
		Quantity float64
	}
	// DepthItemType - тип для зберігання заявок в стакані
	Depth struct {
		symbol            string
		degree            int
		asks              *btree.BTree
		asksSummaQuantity float64
		asksMinMax        *btree.BTree
		bids              *btree.BTree
		bidsSummaQuantity float64
		bidsMinMax        *btree.BTree
		mutex             *sync.Mutex
		LastUpdateID      int64
	}
)

func (i *DepthItem) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItem).Price
}

func (i *DepthItem) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItem).Price
}

func (i *QuantityItem) Less(than btree.Item) bool {
	return i.Quantity < than.(*QuantityItem).Quantity
}

func (i *QuantityItem) Equal(than btree.Item) bool {
	return i.Quantity == than.(*QuantityItem).Quantity
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
	asks.Ascend(func(i btree.Item) bool {
		d.asksSummaQuantity += i.(*DepthItem).Quantity
		d.asksMinMax.ReplaceOrInsert(&QuantityItem{i.(*DepthItem).Quantity, i.(*DepthItem).Price})
		return true
	})
}

// SetBids implements depth_interface.Depths.
func (d *Depth) SetBids(bids *btree.BTree) {
	d.bids = bids
	bids.Ascend(func(i btree.Item) bool {
		d.bidsSummaQuantity += i.(*DepthItem).Quantity
		d.bidsMinMax.ReplaceOrInsert(&QuantityItem{i.(*DepthItem).Quantity, i.(*DepthItem).Price})
		return true
	})
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
	item := d.asks.Get(&DepthItem{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price float64) btree.Item {
	item := d.bids.Get(&DepthItem{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// SetAsk implements depth_interface.Depths.
func (d *Depth) SetAsk(price float64, quantity float64) {
	old := d.asks.Get(&DepthItem{Price: price})
	if old != nil {
		d.asksSummaQuantity -= old.(*DepthItem).Quantity
	}
	d.asks.ReplaceOrInsert(&DepthItem{Price: price, Quantity: quantity})
	d.asksSummaQuantity += quantity
	d.asksMinMax.ReplaceOrInsert(&QuantityItem{quantity, price})
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(price float64, quantity float64) {
	old := d.bids.Get(&DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*DepthItem).Quantity
	}
	d.bids.ReplaceOrInsert(&DepthItem{Price: price, Quantity: quantity})
	d.bidsSummaQuantity += quantity
	d.bidsMinMax.ReplaceOrInsert(&QuantityItem{quantity, price})
}

// DeleteAsk implements depth_interface.Depths.
func (d *Depth) DeleteAsk(price float64) {
	old := d.asks.Get(&DepthItem{Price: price})
	if old != nil {
		d.asksSummaQuantity -= old.(*DepthItem).Quantity
		d.asksMinMax.Delete(&QuantityItem{Quantity: old.(*DepthItem).Quantity})
	}
	d.asks.Delete(&DepthItem{Price: price})
}

// DeleteBid implements depth_interface.Depths.
func (d *Depth) DeleteBid(price float64) {
	old := d.bids.Get(&DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*DepthItem).Quantity
		d.bidsMinMax.Delete(&QuantityItem{Quantity: old.(*DepthItem).Quantity})
	}
	d.bids.Delete(&DepthItem{Price: price})
}

func (d *Depth) GetAsksSummaQuantity() float64 {
	return d.asksSummaQuantity
}

func (d *Depth) GetBidsSummaQuantity() float64 {
	return d.bidsSummaQuantity
}

// RestrictAsk implements depth_interface.Depths.
func (d *Depth) RestrictAsk(price float64) {
	d.asks.Ascend(func(i btree.Item) bool {
		if i.(*DepthItem).Price < price {
			d.asks.Delete(i)
			return false
		}
		return true
	})
}

// RestrictBid implements depth_interface.Depths.
func (d *Depth) RestrictBid(price float64) {
	d.bids.Ascend(func(i btree.Item) bool {
		if i.(*DepthItem).Price > price {
			d.bids.Delete(i)
			return false
		}
		return true
	})
}

// UpdateAsk implements depth_interface.Depths.
func (d *Depth) UpdateAsk(price float64, quantity float64) bool {
	if quantity == 0 {
		d.DeleteAsk(price)
	} else {
		d.SetAsk(price, quantity)
		d.DeleteBid(price)
	}
	return true
}

// UpdateBid implements depth_interface.Depths.
func (d *Depth) UpdateBid(price float64, quantity float64) bool {
	if quantity == 0 {
		d.DeleteBid(price)
	} else {
		d.SetBid(price, quantity)
		d.DeleteAsk(price)
	}
	return true
}

type DepthFilter func(float64) bool

func (d *Depth) getIterator(newTree *btree.BTree, f ...DepthFilter) func(i btree.Item) bool {
	return func(i btree.Item) bool {
		var filter DepthFilter
		pp := i.(*DepthItem)
		quantity := (pp.Quantity / d.asksSummaQuantity) * 100
		if len(f) > 0 {
			filter = f[0]
		} else {
			filter = func(float64) bool { return true }
		}
		if filter(quantity) {
			newTree.ReplaceOrInsert(&DepthItem{
				Price:    pp.Price,
				Quantity: pp.Quantity})
		}
		return true // продовжуємо обхід
	}
}

func (d *Depth) GetFilteredByPercentAsks(f ...DepthFilter) *btree.BTree {
	newTree := btree.New(d.degree)
	if len(f) > 0 {
		d.AskAscend(d.getIterator(newTree, f[0]))
	} else {
		d.AskAscend(d.getIterator(newTree))
	}

	return newTree
}

func (d *Depth) GetFilteredByPercentBids(f ...DepthFilter) *btree.BTree {
	newTree := btree.New(d.degree)
	if len(f) > 0 {
		d.BidDescend(d.getIterator(newTree, f[0]))
	} else {
		d.BidDescend(d.getIterator(newTree))
	}

	return newTree
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

func (d *Depth) AskMin() (min *QuantityItem, err error) {
	if d.asksMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	min = d.asksMinMax.Min().(*QuantityItem)
	return
}

func (d *Depth) AskMax() (max *QuantityItem, err error) {
	if d.asksMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	max = d.asksMinMax.Max().(*QuantityItem)
	return
}

func (d *Depth) BidMin() (min *QuantityItem, err error) {
	if d.bidsMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	min = d.bidsMinMax.Min().(*QuantityItem)
	return
}

func (d *Depth) BidMax() (max *QuantityItem, err error) {
	if d.bidsMinMax.Len() == 0 {
		err = errors.New("asksMinMax is empty")
	}
	max = d.bidsMinMax.Max().(*QuantityItem)
	return
}

// DepthBTree - B-дерево для зберігання стакана заявок
func New(degree int, symbol string) *Depth {
	return &Depth{
		symbol:     symbol,
		degree:     degree,
		asks:       btree.New(degree),
		asksMinMax: btree.New(degree),
		bids:       btree.New(degree),
		bidsMinMax: btree.New(degree),
		mutex:      &sync.Mutex{},
	}
}

func Binance2BookTicker(binanceDepth interface{}) (*DepthItem, error) {
	switch binanceDepth := binanceDepth.(type) {
	case *DepthItem:
		return binanceDepth, nil
	}
	return nil, errors.New("it's not a DepthItemType")
}
