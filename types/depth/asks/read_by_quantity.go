package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaByQuantity(targetSumma items.QuantityType, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantity(targetSumma, depths.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaByQuantityPercent(target float64, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByQuantityPercent(target, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxByQuantity() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxByQuantity(depths.UP)
}
