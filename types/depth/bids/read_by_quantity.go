package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaQuantityByQuantity(targetSumma items.QuantityType, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByQuantity(targetSumma, depths.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaQuantityByQuantityPercent(target float64, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByQuantityPercent(target, depths.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxQuantity() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxQuantity(depths.DOWN)
}
