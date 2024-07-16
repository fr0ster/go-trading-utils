package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depth) GetAsksBidMaxAndSummaByQuantity(targetSummaAsk, targetSummaBid float64, firstMax ...bool) (
	asks,
	bids *types.DepthItem,
	summaAsks,
	summaBids float64) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target float64, item *types.DepthItem, summa *float64) func(i btree.Item) bool {
		buffer := 0.0
		return func(i btree.Item) bool {
			if (*summa + i.(*types.DepthItem).Quantity) < target {
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
	d.GetAsks().Ascend(getIterator(targetSummaAsk, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetSummaBid, bids, &summaBids))
	return
}
func (d *Depth) GetAsksBidMaxAndSummaByQuantityPercent(targetPercentAsk, targetPercentBid float64) (
	asks,
	bids *types.DepthItem,
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
	getIterator := func(targetPercent float64, max, item *types.DepthItem, summa *float64) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			*summa += i.(*types.DepthItem).Quantity
			if (i.(*types.DepthItem).Quantity)*100/max.Quantity >= targetPercent {
				item.Price = i.(*types.DepthItem).Price
				item.Quantity = i.(*types.DepthItem).Quantity
				return false
			} else {
				return true
			}
		}
	}
	asks = &types.DepthItem{}
	bids = &types.DepthItem{}
	d.GetAsks().Ascend(getIterator(targetPercentAsk, maxAsks, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetPercentBid, maxBids, bids, &summaBids))
	return
}
