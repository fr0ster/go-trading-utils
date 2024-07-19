package depth

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Depths) GetTargetPrices(percent float64) (
	priceUp,
	priceDown items.PriceType,
	summaAsks,
	summaBids items.ValueType,
	summaQuantityAsks,
	summaQuantityBids items.QuantityType) {
	upDepthItem, summaAsks, summaQuantityAsks := d.GetAsks().GetMaxAndSummaByValuePercent(percent)
	DownDepthItem, summaBids, summaQuantityBids := d.GetBids().GetMaxAndSummaByValuePercent(percent)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
func (d *Depths) GetLimitPrices() (
	priceUp,
	priceDown items.PriceType,
	summaValueAsks,
	summaValueBids items.ValueType,
	summaQuantityAsks,
	summaQuantityBids items.QuantityType) {

	var (
		askMax *items.DepthItem
		bidMax *items.DepthItem
	)
	_, askMax = d.GetAsks().GetMinMaxByQuantity()
	_, bidMax = d.GetBids().GetMinMaxByQuantity()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	_, summaValueAsks, summaQuantityAsks = d.GetAsks().GetMaxAndSummaByPrice(priceUp)
	_, summaValueBids, summaQuantityBids = d.GetBids().GetMaxAndSummaByPrice(priceDown)
	return
}
