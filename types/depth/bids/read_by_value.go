package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Bids) GetMaxAndSummaValueByPrice(targetPrice items.PriceType, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaValueByPrice(targetPrice, depths.DOWN, firstMax...)
}

// Відбираємо по сумі
func (d *Bids) GetMaxAndSummaValue(targetSumma items.ValueType, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaValue(targetSumma, depths.DOWN, firstMax...)
}

func (d *Bids) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths.DOWN, firstMax...)
}

func (d *Bids) GetMinMaxValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxValue(depths.DOWN)
}
