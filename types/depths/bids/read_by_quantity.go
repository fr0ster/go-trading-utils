package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

// Відбираємо по сумі
func (d *Bids) GetSummaByQuantity(targetSumma items_types.QuantityType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByQuantity(targetSumma, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetSummaByQuantityPercent(target items_types.PricePercentType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByQuantityPercent(target, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByQuantity() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByQuantity(depths_types.DOWN)
}
