package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// func (d *Bids) GetMaxAndSummaValueByPrice(targetPrice items.PriceType, firstMax ...bool) (
// 	item *items.DepthItem,
// 	summaValue items.ValueType,
// 	summaQuantity items.QuantityType) {
// 	return d.tree.GetMaxAndSummaValueByPrice(targetPrice, depths.DOWN, firstMax...)
// }

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaValue(targetSumma items.ValueType, firstMax ...bool) (
	item *items.DepthItem,
	summaValue items.ValueType,
	summaQuantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByValue(targetSumma, depths.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (
	item *items.DepthItem,
	summaValue items.ValueType,
	summaQuantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxByValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxByValue(depths.DOWN)
}
