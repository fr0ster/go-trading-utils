package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaByValue(targetSumma items.ValueType, firstMax ...bool) (
	item *items.DepthItem,
	summaValue items.ValueType,
	summaQuantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByValue(targetSumma, depths.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (
	item *items.DepthItem,
	summaValue items.ValueType,
	summaQuantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxByValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxByValue(depths.UP)
}
