package asks

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetMaxAndSummaByPrice(targetPrice items_types.PriceType, firstMax ...bool) (
	item *items_types.DepthItem,
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetMaxAndSummaByPrice(targetPrice, depths_types.UP, firstMax...)
}

func (d *Asks) GetMinMaxByPrice() (min, max *items_types.DepthItem) {
	return d.tree.GetMinMaxByPrice(depths_types.DOWN)
}
