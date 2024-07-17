package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

// GetAsk implements depth_interface.Depths.
func (d *Depth) GetAsk(price types.PriceType) btree.Item {
	item := d.asks.Get(types.NewDepthItem(price))
	if item == nil {
		return nil
	}
	return item
}

// GetBid implements depth_interface.Depths.
func (d *Depth) GetBid(price types.PriceType) btree.Item {
	item := d.bids.Get(types.NewDepthItem(price))
	if item == nil {
		return nil
	}
	return item
}

// SetAsk implements depth_interface.Depths.
func (d *Depth) SetAsk(price types.PriceType, quantity types.QuantityType) (err error) {
	if old := d.asks.Get(types.NewDepthItem(price)); old != nil {
		d.asksSummaQuantity += quantity - old.(*types.DepthItem).GetQuantity()
	} else {
		d.asksSummaQuantity += quantity
		d.asksCountQuantity++
	}
	d.asks.ReplaceOrInsert(types.NewDepthItem(price, quantity))
	// d.AddAskMinMax(price, quantity)
	// err = d.AddAskNormalized(price, quantity)
	return
}

// SetBid implements depth_interface.Depths.
func (d *Depth) SetBid(price types.PriceType, quantity types.QuantityType) (err error) {
	if old := d.bids.Get(types.NewDepthItem(price)); old != nil {
		d.bidsSummaQuantity += quantity - old.(*types.DepthItem).GetQuantity()
	} else {
		d.bidsSummaQuantity += quantity
		d.bidsCountQuantity++
	}
	d.bids.ReplaceOrInsert(types.NewDepthItem(price, quantity))
	// d.AddBidMinMax(price, quantity)
	// err = d.AddBidNormalized(price, quantity)
	return
}

// DeleteAsk implements depth_interface.Depths.
func (d *Depth) DeleteAsk(price types.PriceType) {
	old := d.asks.Get(types.NewDepthItem(price))
	if old != nil {
		d.asksSummaQuantity -= old.(*types.DepthItem).GetQuantity()
		d.asksCountQuantity--
		// d.DeleteAskMinMax(price, old.(*types.DepthItem).GetQuantity())
		// d.DeleteAskNormalized(price, old.(*types.DepthItem).GetQuantity())
		d.asks.Delete(types.NewDepthItem(price))
	}
}

// DeleteBid implements depth_interface.Depths.
func (d *Depth) DeleteBid(price types.PriceType) {
	old := d.bids.Get(types.NewDepthItem(price))
	if old != nil {
		d.bidsSummaQuantity -= old.(*types.DepthItem).GetQuantity()
		d.asksCountQuantity--
		// d.DeleteBidMinMax(price, old.(*types.DepthItem).GetQuantity())
		// d.DeleteBidNormalized(price, old.(*types.DepthItem).GetQuantity())
		d.bids.Delete(types.NewDepthItem(price))
	}
}

// UpdateAsk implements depth_interface.Depths.
func (d *Depth) UpdateAsk(price types.PriceType, quantity types.QuantityType) bool {
	if quantity == 0 {
		d.DeleteAsk(price)
	} else {
		d.SetAsk(price, quantity)
		d.DeleteBid(price)
	}
	return true
}

// UpdateBid implements depth_interface.Depths.
func (d *Depth) UpdateBid(price types.PriceType, quantity types.QuantityType) bool {
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
