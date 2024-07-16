package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

// Відбираємо по сумі
func (d *Depth) GetAsksBidMaxAndSummaByQuantity(targetSummaAsk, targetSummaBid types.QuantityType, firstMax ...bool) (
	asks,
	bids *types.DepthItem,
	summaAsks,
	summaBids types.QuantityType) {
	var IsFirstMax bool
	if len(firstMax) > 0 {
		IsFirstMax = firstMax[0]
	}
	getIterator := func(target types.QuantityType, item *types.DepthItem, summa *types.QuantityType) func(i btree.Item) bool {
		buffer := types.QuantityType(0.0)
		return func(i btree.Item) bool {
			if (*summa + i.(*types.DepthItem).GetQuantity()) < target {
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
	d.GetAsks().Ascend(getIterator(targetSummaAsk, asks, &summaAsks))
	d.GetBids().Descend(getIterator(targetSummaBid, bids, &summaBids))
	return
}
func (d *Depth) GetAsksBidMaxAndSummaByQuantityPercent(targetPercentAsk, targetPercentBid float64) (
	asks,
	bids *types.DepthItem,
	summaAsks,
	summaBids types.QuantityType,
	err error) {
	maxAsks, err := d.AskMax()
	if err != nil {
		return
	}
	maxBids, err := d.BidMax()
	if err != nil {
		return
	}
	getIterator := func(targetPercent types.QuantityType, max, item *types.DepthItem, summa *types.QuantityType) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			*summa += i.(*types.DepthItem).GetQuantity()
			if (i.(*types.DepthItem).GetQuantity())*100/max.GetQuantity() >= targetPercent {
				item.SetPrice(i.(*types.DepthItem).GetPrice())
				item.SetQuantity(i.(*types.DepthItem).GetQuantity())
				return false
			} else {
				return true
			}
		}
	}
	asks = &types.DepthItem{}
	bids = &types.DepthItem{}
	d.GetAsks().Ascend(getIterator(types.QuantityType(targetPercentAsk), maxAsks, asks, &summaAsks))
	d.GetBids().Descend(getIterator(types.QuantityType(targetPercentBid), maxBids, bids, &summaBids))
	return
}
