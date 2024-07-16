package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

func (d *Depth) GetAsksBidMaxAndSummaByPrice(targetPriceAsk, targetPriceBid types.PriceType, firstMax ...bool) (
	asks,
	bids *types.DepthItem,
	summaAsks,
	summaBids types.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(
		targetPrice types.PriceType,
		item *types.DepthItem,
		summa *types.QuantityType,
		f func(types.PriceType, types.PriceType) bool) func(i btree.Item) bool {
		buffer := types.QuantityType(0.0)
		return func(i btree.Item) bool {
			if f(i.(*types.DepthItem).GetPrice(), targetPrice) {
				buffer += i.(*types.DepthItem).GetQuantity()
				if !IsFirstMax || i.(*types.DepthItem).GetQuantity() >= item.GetQuantity() {
					item.SetPrice(i.(*types.DepthItem).GetPrice())
					item.SetQuantity(i.(*types.DepthItem).GetQuantity())
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
			func(price types.PriceType, target types.PriceType) bool { return price <= target }))
	d.GetBids().Descend(
		getIterator(
			targetPriceBid,
			bids,
			&summaBids,
			func(price types.PriceType, target types.PriceType) bool { return price >= target }))
	return
}
