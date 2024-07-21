package bids

import (
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Bids) GetFiltered(f ...items_types.DepthFilter) (bids *Bids) {
	bids = New(d.Degree(), d.Symbol())
	bids.SetTree(d.tree.GetFiltered(depths_types.DOWN, f...).GetTree())
	return
}

func (d *Bids) GetSummaByPriceRange(
	first,
	last items_types.PriceType,
	f ...items_types.DepthFilter) (
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPriceRange(first, last, f...)
}
