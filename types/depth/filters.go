package depth

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/google/btree"
)

func (d *Depth) getIterator(tree *btree.BTree, summa, max, min *float64, f ...types.DepthFilter) func(i btree.Item) bool {
	return func(i btree.Item) bool {
		var filter types.DepthFilter
		pp := i.(*types.DepthItem)
		if len(f) > 0 {
			filter = f[0]
		} else {
			filter = func(*types.DepthItem) bool { return true }
		}
		if filter(pp) {
			tree.ReplaceOrInsert(&types.DepthItem{
				Price:    pp.Price,
				Quantity: pp.Quantity})
			if summa != nil {
				*summa += pp.Quantity
			}
			if max != nil {
				if *max < pp.Quantity {
					*max = pp.Quantity
				}
			}
			if min != nil {
				if *min > pp.Quantity || *min == 0 {
					*min = pp.Quantity
				}
			}
		}
		return true // продовжуємо обхід
	}
}

func (d *Depth) GetFilteredByPercentAsks(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min float64) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.AskAscend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.AskAscend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depth) GetFilteredByPercentBids(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min float64) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.BidDescend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.BidDescend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depth) GetSummaOfAsksFromRange(first, last float64, f ...types.DepthFilter) (askSumma, max, min float64) {
	var filter types.DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*types.DepthItem) bool { return true }
	}
	d.GetAsks().DescendRange(&types.DepthItem{Price: last}, &types.DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*types.DepthItem)) {
			askSumma += i.(*types.DepthItem).Quantity
			if max < i.(*types.DepthItem).Quantity {
				max = i.(*types.DepthItem).Quantity
			}
			if min > i.(*types.DepthItem).Quantity || min == 0 {
				min = i.(*types.DepthItem).Quantity
			}
		}
		return true
	})
	return
}

func (d *Depth) GetSummaOfBidsFromRange(first, last float64, f ...types.DepthFilter) (bidSumma, max, min float64) {
	var filter types.DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*types.DepthItem) bool { return true }
	}
	d.GetBids().AscendRange(&types.DepthItem{Price: last}, &types.DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*types.DepthItem)) {
			bidSumma += i.(*types.DepthItem).Quantity
			if max < i.(*types.DepthItem).Quantity {
				max = i.(*types.DepthItem).Quantity
			}
			if min > i.(*types.DepthItem).Quantity || min == 0 {
				min = i.(*types.DepthItem).Quantity
			}
		}
		return true
	})
	return
}

func (d *Depth) GetPercentToTarget() float64 {
	return d.percentToTarget
}

// RestrictAskUp implements depth_interface.Depths.
func (d *Depth) RestrictAskUp(price float64) {
	prices := make([]float64, 0)
	d.asks.AscendGreaterOrEqual(&types.DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.asks.Delete(&types.DepthItem{Price: p})
	}
}

// RestrictBidUp implements depth_interface.Depths.
func (d *Depth) RestrictBidUp(price float64) {
	prices := make([]float64, 0)
	d.bids.AscendGreaterOrEqual(&types.DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.bids.Delete(&types.DepthItem{Price: p})
	}
}

// RestrictAskDown implements depth_interface.Depths.
func (d *Depth) RestrictAskDown(price float64) {
	prices := make([]float64, 0)
	d.asks.DescendLessOrEqual(&types.DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.asks.Delete(&types.DepthItem{Price: p})
	}
}

// RestrictBidDown implements depth_interface.Depths.
func (d *Depth) RestrictBidDown(price float64) {
	prices := make([]float64, 0)
	d.bids.DescendLessOrEqual(&types.DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.bids.Delete(&types.DepthItem{Price: p})
	}
}
