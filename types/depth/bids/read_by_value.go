package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Bids) GetMaxAndSummaValueByPrice(targetPrice items.PriceType, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValueByPrice(targetPrice, depths.DOWN, firstMax...)
}

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaValue(targetSumma items.ValueType, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValue(targetSumma, depths.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaValueByPercent(target float64, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValueByPercent(target, depths.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxValue(depths.DOWN)
}
