package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaByQuantity(targetSumma items_types.QuantityType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantity(targetSumma, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaByQuantityPercent(target float64, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantityPercent(target, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByQuantity() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByQuantity(depths_types.DOWN)
}
