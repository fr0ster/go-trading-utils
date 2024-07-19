package bids

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Clear implements depth_interface.Depths.
func (d *Bids) Clear() {
	d.tree.Clear()
}

func (d *Bids) GetMiddleQuantity() types.QuantityType {
	return d.tree.GetMiddleQuantity()
}

func (d *Bids) GetMiddleValue() types.ValueType {
	return d.tree.GetMiddleValue()
}

func (d *Bids) GetStandardDeviation() float64 {
	return d.tree.GetStandardDeviation()
}
