package depth

import (
	"github.com/fr0ster/go-trading-utils/types/depth/depths"
	items "github.com/fr0ster/go-trading-utils/types/depth/items"
)

func (d *Depths) GetTargetPrices(percent float64) (priceUp, priceDown items.PriceType, summaAsks, summaBids items.QuantityType) {
	upDepthItem, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(
		d.GetAsks().GetDepths().GetSummaQuantity()*items.QuantityType(percent)/100, depths.UP)
	DownDepthItem, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(
		d.GetBids().GetDepths().GetSummaQuantity()*items.QuantityType(percent)/100, depths.DOWN)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
func (d *Depths) GetLimitPrices() (priceUp, priceDown items.PriceType, summaAsks, summaBids items.QuantityType) {

	var (
		askMax *items.DepthItem
		bidMax *items.DepthItem
	)
	_, askMax = d.GetAsks().GetDepths().GetMinMaxQuantity(depths.UP)
	_, bidMax = d.GetBids().GetDepths().GetMinMaxQuantity(depths.DOWN)
	priceUp = askMax.GetPrice()
	priceDown = bidMax.GetPrice()
	_, summaAsks = d.GetAsks().GetDepths().GetMaxAndSummaByPrice(priceUp, depths.UP)
	_, summaBids = d.GetBids().GetDepths().GetMaxAndSummaByPrice(priceDown, depths.DOWN)
	return
}
