package depth

func (d *Depth) GetTargetPrices(percent float64) (priceUp, priceDown float64) {
	upDepthItem, DownDepthItem, _, _ := d.GetAsksBidFirstMaxAndSumma(
		d.GetAsksSummaQuantity()*percent/100,
		d.GetBidsSummaQuantity()*percent/100,
		true,
	)
	priceUp = upDepthItem.Price
	priceDown = DownDepthItem.Price
	return
}
