package bids

import (
	depths "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Bids) GetMaxAndSummaQuantityByPrice(targetPrice items.PriceType, firstMax ...bool) (item *items.DepthItem, quantity items.QuantityType) {
	return d.tree.GetMaxAndSummaQuantityByPrice(targetPrice, depths.DOWN, firstMax...)
}
