package depth

import (
	"github.com/google/btree"
)

func (d *Depth) getIterator(tree *btree.BTree, summa, max, min *float64, f ...DepthFilter) func(i btree.Item) bool {
	return func(i btree.Item) bool {
		var filter DepthFilter
		pp := i.(*DepthItem)
		if len(f) > 0 {
			filter = f[0]
		} else {
			filter = func(*DepthItem) bool { return true }
		}
		if filter(pp) {
			tree.ReplaceOrInsert(&DepthItem{
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

func (d *Depth) GetFilteredByPercentAsks(f ...DepthFilter) (tree *btree.BTree, summa, max, min float64) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.AskAscend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.AskAscend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depth) GetFilteredByPercentBids(f ...DepthFilter) (tree *btree.BTree, summa, max, min float64) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.BidDescend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.BidDescend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depth) GetSummaOfAsksFromRange(first, last float64, f ...DepthFilter) (askSumma, max, min float64) {
	var filter DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*DepthItem) bool { return true }
	}
	d.GetAsks().DescendRange(&DepthItem{Price: last}, &DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*DepthItem)) {
			askSumma += i.(*DepthItem).Quantity
			if max < i.(*DepthItem).Quantity {
				max = i.(*DepthItem).Quantity
			}
			if min > i.(*DepthItem).Quantity || min == 0 {
				min = i.(*DepthItem).Quantity
			}
		}
		return true
	})
	return
}

func (d *Depth) GetSummaOfBidsFromRange(first, last float64, f ...DepthFilter) (bidSumma, max, min float64) {
	var filter DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*DepthItem) bool { return true }
	}
	d.GetBids().AscendRange(&DepthItem{Price: last}, &DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*DepthItem)) {
			bidSumma += i.(*DepthItem).Quantity
			if max < i.(*DepthItem).Quantity {
				max = i.(*DepthItem).Quantity
			}
			if min > i.(*DepthItem).Quantity || min == 0 {
				min = i.(*DepthItem).Quantity
			}
		}
		return true
	})
	return
}

func (d *Depth) GetPercentToTarget() float64 {
	return d.percentRoTarget
}

func (d *Depth) GetPercentToLimit() float64 {
	return d.percentToLimit
}

// RestrictAskUp implements depth_interface.Depths.
func (d *Depth) RestrictAskUp(price float64) {
	prices := make([]float64, 0)
	d.asks.AscendGreaterOrEqual(&DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.asks.Delete(&DepthItem{Price: p})
	}
}

// RestrictBidUp implements depth_interface.Depths.
func (d *Depth) RestrictBidUp(price float64) {
	prices := make([]float64, 0)
	d.bids.AscendGreaterOrEqual(&DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.bids.Delete(&DepthItem{Price: p})
	}
}

// RestrictAskDown implements depth_interface.Depths.
func (d *Depth) RestrictAskDown(price float64) {
	prices := make([]float64, 0)
	d.asks.DescendLessOrEqual(&DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.asks.Delete(&DepthItem{Price: p})
	}
}

// RestrictBidDown implements depth_interface.Depths.
func (d *Depth) RestrictBidDown(price float64) {
	prices := make([]float64, 0)
	d.bids.DescendLessOrEqual(&DepthItem{Price: price}, func(i btree.Item) bool {
		prices = append(prices, i.(*DepthItem).Price)
		return true
	})
	for _, p := range prices {
		d.bids.Delete(&DepthItem{Price: p})
	}
}
