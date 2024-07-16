package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

func (d *Depth) GetAsksBidMaxAndSummaByPrice(targetPriceAsk, targetPriceBid float64, firstMax ...bool) (
	asks,
	bids *types.DepthItem,
	summaAsks,
	summaBids float64) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice float64,
		item *types.DepthItem,
		summa *float64,
		f func(float64, float64) bool) func(i btree.Item) bool {
		buffer := 0.0
		return func(i btree.Item) bool {
			if f(i.(*types.DepthItem).Price, targetPrice) {
				buffer += i.(*types.DepthItem).Quantity
				if !IsFirstMax || i.(*types.DepthItem).Quantity >= item.Quantity {
					item.Price = i.(*types.DepthItem).Price
					item.Quantity = i.(*types.DepthItem).Quantity
					*summa = buffer
				}
				return true
			} else {
				return false
			}
		}
	}
	asks = &types.DepthItem{}
	bids = &types.DepthItem{}
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
