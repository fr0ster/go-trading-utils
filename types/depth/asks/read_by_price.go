package asks

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetMaxAndSummaQuantityByPrice(targetPrice items.PriceType, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByPrice(targetPrice, depths.UP, firstMax...)
}
