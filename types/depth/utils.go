package depth

func (d *Depth) GetTargetPrices(percent float64) (priceUp, priceDown, summaAsks, summaBids float64) {
	upDepthItem, DownDepthItem, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByQuantity(
		d.GetAsksSummaQuantity()*percent/100,
		d.GetBidsSummaQuantity()*percent/100,
		true,
	)
	priceUp = upDepthItem.GetPrice()
	priceDown = DownDepthItem.GetPrice()
	return
}
