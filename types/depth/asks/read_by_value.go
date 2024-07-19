package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetMaxAndSummaValueByPrice(targetPrice items.PriceType, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaValueByPrice(targetPrice, depths.UP, firstMax...)
}

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaValue(targetSumma items.ValueType, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaValue(targetSumma, depths.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaByValuePercent(target float64, firstMax ...bool) (item *items.DepthItem, summa items.ValueType) {
	return d.tree.GetMaxAndSummaByValuePercent(target, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxValue(depths.UP)
}
