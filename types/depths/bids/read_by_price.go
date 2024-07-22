package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

func (d *Bids) GetSummaByPrice(targetPrice items_types.PriceType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPrice(targetPrice, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetSummaByPricePercent(targetPrice items_types.PricePercentType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPricePercent(targetPrice, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByPrice() (min, max *items_types.DepthItem, err error) {
	return d.tree.GetMinMaxByPrice(depths_types.DOWN)
}
