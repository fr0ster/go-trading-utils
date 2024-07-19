package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaValue(targetSumma items_types.ValueType, firstMax ...bool) (
	item *items_types.DepthItem,
	summaValue items_types.ValueType,
	summaQuantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByValue(targetSumma, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (
	item *items_types.DepthItem,
	summaValue items_types.ValueType,
	summaQuantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths_types.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByValue() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByValue(depths_types.DOWN)
}
