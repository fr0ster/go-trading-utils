package depth

import (
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Depths) GetTargetPrices(percent float64) (priceUp, priceDown items.PriceType, summaAsks, summaBids items.QuantityType) {
	upDepthItem, summaAsks := d.GetAsks().GetMaxAndSummaQuantityByQuantity(
		d.GetAsks().GetSummaQuantity() * items.QuantityType(percent) / 100)
	DownDepthItem, summaBids := d.GetBids().GetMaxAndSummaQuantityByQuantity(
		d.GetBids().GetSummaQuantity() * items.QuantityType(percent) / 100)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
func (d *Depths) GetLimitPrices() (priceUp, priceDown items.PriceType, summaAsks, summaBids items.QuantityType) {

	var (
		askMax *items.DepthItem
		bidMax *items.DepthItem
	)
	_, askMax = d.GetAsks().GetMinMaxQuantity()
	_, bidMax = d.GetBids().GetMinMaxQuantity()
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	_, summaAsks = d.GetAsks().GetMaxAndSummaQuantityByPrice(priceUp)
	_, summaBids = d.GetBids().GetMaxAndSummaQuantityByPrice(priceDown)
	return
}
