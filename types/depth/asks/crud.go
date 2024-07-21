package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (a *Asks) Get(item *items_types.Ask) *items_types.Ask {
	if val := a.tree.Get((*items_types.DepthItem)(item)); val != nil {
		return (*items_types.Ask)(val)
	} else {
		return nil
	}
}

func (a *Asks) Set(item *items_types.Ask) (err error) {
	return a.tree.Set((*items_types.DepthItem)(item))
}

func (a *Asks) Delete(item *items_types.Ask) {
	a.tree.Delete(item.GetDepthItem())
}

func (a *Asks) Update(item *items_types.Ask) bool {
	return a.tree.Update((*items_types.DepthItem)(item))
}

// Count implements depth_interface.Depths.
func (d *Asks) Count() int {
	return d.tree.Count()
}

func (d *Asks) GetSummaQuantity() items_types.QuantityType {
	return d.tree.GetSummaQuantity()
}

func (d *Asks) GetSummaValue() items_types.ValueType {
	return d.tree.GetSummaValue()
}

func (d *Asks) GetMiddleQuantity() items_types.QuantityType {
	return d.tree.GetMiddleQuantity()
}

func (d *Asks) GetMiddleValue() items_types.ValueType {
	return d.tree.GetMiddleValue()
}

func (d *Asks) GetMinPrice() (min *items_types.DepthItem, err error) {
	return d.tree.GetMinPrice()
}

func (d *Asks) GetMaxPrice() (max *items_types.DepthItem, err error) {
	return d.tree.GetMaxPrice()
}

func (d *Asks) GetDeltaPrice() (delta items_types.PriceType, err error) {
	return d.tree.GetDeltaPrice()
}

func (d *Asks) GetDeltaPricePercent() (delta items_types.PricePercentType, err error) {
	return d.tree.GetDeltaPricePercent(depths_types.UP)
}

func (d *Asks) GetStandardDeviation() float64 {
	return d.tree.GetStandardDeviation()
}

func (d *Asks) NextPriceUp(percent items_types.PricePercentType) items_types.PriceType {
	return d.tree.NextPriceUp(percent)
}
