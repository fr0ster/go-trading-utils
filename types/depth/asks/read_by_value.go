package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetMaxAndSummaValueByPrice(targetPrice items.PriceType, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValueByPrice(targetPrice, depths.UP, firstMax...)
}

// Відбираємо по сумі
func (d *Asks) GetMaxAndSummaValue(targetSumma items.ValueType, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValue(targetSumma, depths.UP, firstMax...)
}

func (d *Asks) GetMaxAndSummaValueByPercent(target float64, firstMax ...bool) (
	item *types.DepthItem,
	summaValue types.ValueType,
	summaQuantity types.QuantityType) {
	return d.tree.GetMaxAndSummaValueByPercent(target, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxValue() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxValue(depths.UP)
}
