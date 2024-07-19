package depth

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Depths) GetTargetPrices(percent float64) (
	priceUp,
	priceDown items_types.PriceType,
	summaAsks,
	summaBids items_types.ValueType,
	summaQuantityAsks,
	summaQuantityBids items_types.QuantityType) {
	upDepthItem, summaAsks, summaQuantityAsks := d.GetAsks().GetMaxAndSummaByValuePercent(percent)
	DownDepthItem, summaBids, summaQuantityBids := d.GetBids().GetMaxAndSummaByValuePercent(percent)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
func (d *Depths) GetLimitPrices() (
	priceUp,
	priceDown items_types.PriceType,
	summaValueAsks,
	summaValueBids items_types.ValueType,
	summaQuantityAsks,
	summaQuantityBids items_types.QuantityType) {

	var (
		askMax *items_types.DepthItem
		bidMax *items_types.DepthItem
	)
	_, askMax = d.GetAsks().GetMinMaxByQuantity()
	_, bidMax = d.GetBids().GetMinMaxByQuantity()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	_, summaValueAsks, summaQuantityAsks = d.GetAsks().GetMaxAndSummaByPrice(priceUp)
	_, summaValueBids, summaQuantityBids = d.GetBids().GetMaxAndSummaByPrice(priceDown)
	return
}
