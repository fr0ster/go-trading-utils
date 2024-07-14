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
		buffer := 0.0
		return func(i btree.Item) bool {
			if (*summa + i.(*DepthItem).Quantity) < target {
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
	d.GetAsks().Ascend(getIterator(targetSummaAsk, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetSummaBid, bids, &summaBids))
	return
}
func (d *Depth) GetAsksBidMaxAndSummaByQuantityPercent(targetPercentAsk, targetPercentBid float64) (
	asks,
	bids *DepthItem,
	summaAsks,
	summaBids float64,
	err error) {
	maxAsks, err := d.AskMax()
	if err != nil {
		return
	}
	maxBids, err := d.BidMax()
	if err != nil {
		return
	}
	getIterator := func(targetPercent float64, max, item *DepthItem, summa *float64) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			*summa += i.(*DepthItem).Quantity
			if (i.(*DepthItem).Quantity)*100/max.Quantity > targetPercent {
				item.Price = i.(*DepthItem).Price
				item.Quantity = i.(*DepthItem).Quantity
				return false
			} else {
				return true
			}
		}
	}
	asks = &DepthItem{}
	bids = &DepthItem{}
	d.GetAsks().Ascend(getIterator(targetPercentAsk, maxAsks, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetPercentBid, maxBids, bids, &summaBids))
	return
}
