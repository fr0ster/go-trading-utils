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
	upDepthItem, summaAsks, summaQuantityAsks := d.GetAsks().GetMaxAndSummaValueByPercent(percent)
	DownDepthItem, summaBids, summaQuantityBids := d.GetBids().GetMaxAndSummaValueByPercent(percent)
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
	_, askMax = d.GetAsks().GetMinMaxQuantity()
	_, bidMax = d.GetBids().GetMinMaxQuantity()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	_, summaValueAsks, summaQuantityAsks = d.GetAsks().GetMaxAndSummaValueByPrice(priceUp)
	_, summaValueBids, summaQuantityBids = d.GetBids().GetMaxAndSummaValueByPrice(priceDown)
	return
}
