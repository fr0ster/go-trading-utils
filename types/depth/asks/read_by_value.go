package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaByValue(targetSumma items_types.ValueType, firstMax ...bool) (
	item *items_types.DepthItem,
	summaValue items_types.ValueType,
	summaQuantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByValue(targetSumma, depths_types.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (
	item *items_types.DepthItem,
	summaValue items_types.ValueType,
	summaQuantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths_types.UP, firstMax...)
}

func (d *Asks) GetMinMaxByValue() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByValue(depths_types.UP)
}
