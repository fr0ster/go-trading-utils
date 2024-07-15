package depth

import (
	"github.com/google/btree"
)

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
		d.asksCountQuantity--
		d.DeleteAskMinMax(old.(*DepthItem).Quantity, old.(*DepthItem).Price)
	}
	item := &DepthItem{Price: price, Quantity: quantity}
	d.asks.ReplaceOrInsert(item)
	d.asksSummaQuantity += quantity
	d.asksCountQuantity++
	d.AddAskMinMax(price, quantity)
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(price float64, quantity float64) {
	old := d.bids.Get(&DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*DepthItem).Quantity
		d.bidsCountQuantity--
		d.DeleteBidMinMax(old.(*DepthItem).Quantity, old.(*DepthItem).Price)
	}
	item := &DepthItem{Price: price, Quantity: quantity}
	d.bids.ReplaceOrInsert(item)
	d.bidsSummaQuantity += quantity
	d.bidsCountQuantity++
	d.AddBidMinMax(price, quantity)
}

// DeleteAsk implements depth_interface.Depths.
func (d *Depth) DeleteAsk(price float64) {
	old := d.asks.Get(&DepthItem{Price: price})
	if old != nil {
		d.asksSummaQuantity -= old.(*DepthItem).Quantity
		d.DeleteAskMinMax(price, old.(*DepthItem).Quantity)
	}
	d.asks.Delete(&DepthItem{Price: price})
}

// DeleteBid implements depth_interface.Depths.
func (d *Depth) DeleteBid(price float64) {
	old := d.bids.Get(&DepthItem{Price: price})
	if old != nil {
		d.bidsSummaQuantity -= old.(*DepthItem).Quantity
		d.DeleteBidMinMax(price, old.(*DepthItem).Quantity)
	}
	d.bids.Delete(&DepthItem{Price: price})
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
