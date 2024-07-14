package depth

import (
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depth) GetAsksBidMaxAndSummaByQuantity(targetSummaAsk, targetSummaBid float64, firstMax ...bool) (
	asks,
	bids *DepthItem,
	summaAsks,
	summaBids float64) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target float64, item *DepthItem, summa *float64) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			if (*summa+i.(*DepthItem).Quantity) < target && (!IsFirstMax || i.(*DepthItem).Price > item.Price) {
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
	d.GetAsks().Ascend(getIterator(targetSummaAsk, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetSummaBid, bids, &summaBids))
	return
}
