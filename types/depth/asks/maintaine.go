package asks

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Clear implements depth_interface.Depths.
func (d *Asks) Clear() {
	d.tree.Clear()
}

func (d *Asks) GetMiddleQuantity() types.QuantityType {
	return d.tree.GetMiddleQuantity()
}

func (d *Asks) GetMiddleValue() types.ValueType {
	return d.tree.GetMiddleValue()
}

func (d *Asks) GetStandardDeviation() float64 {
	return d.tree.GetStandardDeviation()
}
