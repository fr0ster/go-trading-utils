package asks

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Asks) GetFiltered(f ...items_types.DepthFilter) (asks *Asks) {
	asks = New(d.Degree(), d.Symbol(), d.TargetPercent(), d.LimitDepth(), d.ExpBase(), d.RateStream())
	asks.SetTree(d.tree.GetFiltered(f...).GetTree())
	return
}

func (d *Asks) GetSummaByPriceRange(
	first,
	last items_types.PriceType,
	f ...items_types.DepthFilter) (
	value items_types.ValueType,
	quantity items_types.QuantityType) {
	return d.tree.GetSummaByPriceRange(first, last, f...)
}
