package depth

import items "github.com/fr0ster/go-trading-utils/types/depth/items"

func (d *Depths) GetTargetPrices(percent float64) (priceUp, priceDown items.PriceType, summaAsks, summaBids items.QuantityType) {
	upDepthItem, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(
		d.GetAsks().GetDepths().GetSummaQuantity()*items.QuantityType(percent)/100, true)
	DownDepthItem, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(
		d.GetBids().GetDepths().GetSummaQuantity()*items.QuantityType(percent)/100, false)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
