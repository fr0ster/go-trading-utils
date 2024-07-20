package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Bids) GetSummaByPrice(targetPrice items_types.PriceType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPrice(targetPrice, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetSummaByPricePercent(targetPrice float64, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPricePercent(targetPrice, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByPrice() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByPrice(depths_types.DOWN)
}
