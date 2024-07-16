package depth

import "github.com/fr0ster/go-trading-utils/types/depth/types"

func (d *Depth) GetTargetPrices(percent float64) (priceUp, priceDown types.PriceType, summaAsks, summaBids types.QuantityType) {
	upDepthItem, DownDepthItem, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByQuantity(
		d.GetAsksSummaQuantity()*types.QuantityType(percent)/100,
		d.GetBidsSummaQuantity()*types.QuantityType(percent)/100,
		true,
	)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
