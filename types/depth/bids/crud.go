package bids

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (a *Bids) Get(item *items_types.Bid) *items_types.Bid {
	if val := a.tree.Get((*items_types.DepthItem)(item)); val != nil {
		return (*items_types.Bid)(val)
	} else {
		return nil
	}
}

func (a *Bids) Set(item *items_types.Bid) (err error) {
	return a.tree.Set((*items_types.DepthItem)(item))
}

func (a *Bids) Delete(item *items_types.Bid) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Bids) Update(item *items_types.Bid) bool {
	return a.tree.Update((*items_types.DepthItem)(item))
}

// Count implements depth_interface.Depths.
func (d *Bids) Count() int {
	return d.tree.Count()
}

func (d *Bids) GetSummaQuantity() items_types.QuantityType {
	return d.tree.GetSummaQuantity()
}

func (d *Bids) GetSummaValue() items_types.ValueType {
	return d.tree.GetSummaValue()
}

func (d *Bids) GetMiddleQuantity() items_types.QuantityType {
	return d.tree.GetMiddleQuantity()
}

func (d *Bids) GetMiddleValue() items_types.ValueType {
	return d.tree.GetMiddleValue()
}

func (d *Bids) GetStandardDeviation() float64 {
	return d.tree.GetStandardDeviation()
}
