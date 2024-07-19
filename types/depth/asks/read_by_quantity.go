package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaQuantityByQuantity(targetSumma items.QuantityType, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByQuantity(targetSumma, depths.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaQuantityByQuantityPercent(target float64, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByQuantityPercent(target, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxQuantity() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxQuantity(depths.UP)
}
