package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
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

func (d *Bids) GetMinPrice() (min *items_types.DepthItem, err error) {
	return d.tree.GetMinPrice()
}

func (d *Bids) GetMaxPrice() (max *items_types.DepthItem, err error) {
	return d.tree.GetMaxPrice()
}

func (d *Bids) GetDeltaPrice() (delta items_types.PriceType, err error) {
	return d.tree.GetDeltaPrice()
}

func (d *Bids) GetDeltaPricePercent() (delta items_types.PricePercentType, err error) {
	return d.tree.GetDeltaPricePercent(depths_types.DOWN)
}

func (d *Bids) GetStandardDeviation() float64 {
	return d.tree.GetStandardDeviation()
}

func (d *Bids) NextPriceDown(percent items_types.PricePercentType) items_types.PriceType {
	return d.tree.NextPriceDown(percent)
}
