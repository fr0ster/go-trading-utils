package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetMaxAndSummaByPrice(targetPrice items.PriceType, firstMax ...bool) (
	item *items.DepthItem,
	value items.ValueType,
	quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaByPrice(targetPrice, depths.UP, firstMax...)
}

func (d *Asks) GetMinMaxByPrice() (min, max *items.DepthItem) {
	return d.tree.GetMinMaxByPrice(depths.DOWN)
}
