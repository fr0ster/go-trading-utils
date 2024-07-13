package depth

func (d *Depth) GetTargetPrices(percent float64) (priceUp, priceDown float64) {
	upDepthItem, DownDepthItem, _, _ := d.GetTargetAsksBidPrice(
		d.GetAsksSummaQuantity()*percent/100,
		d.GetBidsSummaQuantity()*percent/100,
	)
	upDepthItem = d.GetAsksMaxUpToPrice(upDepthItem.Price)
	DownDepthItem = d.GetBidsMaxDownToPrice(DownDepthItem.Price)
	priceUp = upDepthItem.Price
	priceDown = DownDepthItem.Price
	return
}
