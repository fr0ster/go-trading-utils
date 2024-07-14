package depth

import (
	"github.com/google/btree"
)

// Відбираємо по ціні
// func (d *Depth) GetAsksMaxAndSummaUp(price ...float64) (limit *DepthItem, summa float64) {
// 	limit = &DepthItem{}
// 	d.GetAsks().Ascend(func(i btree.Item) bool {
// 		if len(price) > 0 && i.(*DepthItem).Price <= price[0] {
// 			summa += i.(*DepthItem).Quantity
// 			if limit.Quantity < i.(*DepthItem).Quantity {
// 				limit.Quantity = i.(*DepthItem).Quantity
// 				limit.Price = i.(*DepthItem).Price
// 			}
// 			return true
// 		} else {
// 			return false
// 		}
// 	})
// 	return
// }

// func (d *Depth) GetBidsMaxAndSummaDown(price ...float64) (limit *DepthItem, summa float64) {
// 	limit = &DepthItem{}
// 	d.GetBids().Descend(func(i btree.Item) bool {
// 		if len(price) > 0 && i.(*DepthItem).Price >= price[0] {
// 			summa += i.(*DepthItem).Quantity
// 			if limit.Quantity < i.(*DepthItem).Quantity {
// 				limit.Quantity = i.(*DepthItem).Quantity
// 				limit.Price = i.(*DepthItem).Price
// 			}
// 			return true
// 		} else {
// 			return false
// 		}
// 	})
// 	return
// }

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
		fq func(float64, float64) bool) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if fq(i.(*DepthItem).Price, targetPrice) && (!IsFirstMax || i.(*DepthItem).Quantity >= item.Quantity) {
				item.Price = i.(*DepthItem).Price
				item.Quantity = i.(*DepthItem).Quantity
				*summa += i.(*DepthItem).Quantity
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
			func(price float64, target float64) bool { return price < target }))
	d.GetBids().Descend(
		getIterator(
			targetPriceBid,
			bids,
			&summaBids,
			func(price float64, target float64) bool { return price > target }))
	return
}
