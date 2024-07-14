package depth

import (
	"github.com/google/btree"
)

func (d *Depth) GetAsksBidMaxAndSummaByPrice(targetPriceAsk, targetPriceBid float64, firstMax ...bool) (
	asks,
	bids *DepthItem,
	summaAsks,
	summaBids float64) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice float64,
		item *DepthItem,
		summa *float64,
		f func(float64, float64) bool) func(i btree.Item) bool {
		buffer := 0.0
		return func(i btree.Item) bool {
			if f(i.(*DepthItem).Price, targetPrice) {
				buffer += i.(*DepthItem).Quantity
				if !IsFirstMax || i.(*DepthItem).Quantity >= item.Quantity {
					item.Price = i.(*DepthItem).Price
					item.Quantity = i.(*DepthItem).Quantity
					*summa = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	asks = &DepthItem{}
	bids = &DepthItem{}
	d.GetAsks().Ascend(
		getIterator(
			targetPriceAsk,
			asks,
			&summaAsks,
			func(price float64, target float64) bool { return price <= target }))
	d.GetBids().Descend(
		getIterator(
			targetPriceBid,
			bids,
			&summaBids,
			func(price float64, target float64) bool { return price >= target }))
	return
}
