package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaByQuantity(targetSumma items.QuantityType, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantity(targetSumma, depths.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaByQuantityPercent(target float64, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantityPercent(target, depths.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByQuantity() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxByQuantity(depths.DOWN)
}
