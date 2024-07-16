package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

// GetAsk implements depth_interface.Depths.
func (d *Depth) GetAsk(price float64) btree.Item {
	item := d.asks.Get(&types.DepthItem{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price float64) btree.Item {
	item := d.bids.Get(&types.DepthItem{Price: price})
	if item == nil {
		return nil
	}
	return item
}

// SetAsk implements depth_interface.Depths.
func (d *Depth) SetAsk(price float64, quantity float64) (err error) {
	old := d.asks.Get(&types.DepthItem{Price: price})
	if old != nil {
		d.asksSummaQuantity -= old.(*types.DepthItem).Quantity
		d.asksCountQuantity--
		d.DeleteAskMinMax(old.(*types.DepthItem).Quantity, old.(*types.DepthItem).Price)
	}
	item := &types.DepthItem{Price: price, Quantity: quantity}
	d.asks.ReplaceOrInsert(item)
	d.asksSummaQuantity += quantity
	d.asksCountQuantity++
	d.AddAskMinMax(price, quantity)
	err = d.AddAskNormalized(price, quantity)
	return
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(price float64, quantity float64) (err error) {
	old := d.bids.Get(&types.DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*types.DepthItem).Quantity
		d.bidsCountQuantity--
		d.DeleteBidMinMax(old.(*types.DepthItem).Quantity, old.(*types.DepthItem).Price)
	}
	item := &types.DepthItem{Price: price, Quantity: quantity}
	d.bids.ReplaceOrInsert(item)
	d.bidsSummaQuantity += quantity
	d.bidsCountQuantity++
	d.AddBidMinMax(price, quantity)
	err = d.AddBidNormalized(price, quantity)
	return
}

// DeleteAsk implements depth_interface.Depths.
func (d *Depth) DeleteAsk(price float64) {
	old := d.asks.Get(&types.DepthItem{Price: price})
	if old != nil {
		d.asksSummaQuantity -= old.(*types.DepthItem).Quantity
		d.DeleteAskMinMax(price, old.(*types.DepthItem).Quantity)
		d.DeleteAskNormalized(price, old.(*types.DepthItem).Quantity)
	}
	d.asks.Delete(&types.DepthItem{Price: price})
}

// DeleteBid implements depth_interface.Depths.
func (d *Depth) DeleteBid(price float64) {
	old := d.bids.Get(&types.DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*types.DepthItem).Quantity
		d.DeleteBidMinMax(price, old.(*types.DepthItem).Quantity)
		d.DeleteBidNormalized(price, old.(*types.DepthItem).Quantity)
	}
	d.bids.Delete(&types.DepthItem{Price: price})
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

// AskCount implements depth_interface.Depths.
func (d *Depth) AskCount() int {
	return d.asksCountQuantity
}

// BidCount implements depth_interface.Depths.
func (d *Depth) BidCount() int {
	return d.bidsCountQuantity
}
