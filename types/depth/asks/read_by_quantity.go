package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Відбираємо по сумі
func (d *Asks) GetSummaByQuantity(targetSumma items_types.QuantityType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByQuantity(targetSumma, depths_types.UP, firstMax...)
}

func (d *Asks) GetSummaByQuantityPercent(target items_types.PricePercentType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByQuantityPercent(target, depths_types.UP, firstMax...)
}

func (d *Asks) GetMinMaxByQuantity() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByQuantity(depths_types.UP)
}
