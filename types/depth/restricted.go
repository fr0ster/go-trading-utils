package depth

import (
	"github.com/google/btree"
)

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

func (d *Depth) GetTargetAsksBidPrice(targetSummaAsk, targetSummaBid float64) (asks, bids *DepthItem, summaAsks, summaBids float64) {
	getIterator := func(target float64, item *DepthItem, summaOut *float64) func(i btree.Item) bool {
		summa := 0.0
		return func(i btree.Item) bool {
			summa += i.(*DepthItem).Quantity
			if summa < target {
				item.Price = i.(*DepthItem).Price
				item.Quantity = i.(*DepthItem).Quantity
				*summaOut = summa
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

func (d *Depth) GetAsksSumma(price ...float64) (summa float64) {
	d.GetAsks().Ascend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price <= price[0] {
			summa += i.(*DepthItem).Quantity
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetBidsSumma(price ...float64) (summa float64) {
	d.GetBids().Descend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price >= price[0] {
			summa += i.(*DepthItem).Quantity
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetAsksMaxUpToPrice(price ...float64) (limit *DepthItem) {
	limit = &DepthItem{}
	d.GetAsks().Ascend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price <= price[0] {
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetBidsMaxDownToPrice(price ...float64) (limit *DepthItem) {
	limit = &DepthItem{}
	d.GetBids().Descend(func(i btree.Item) bool {
		if len(price) > 0 && i.(*DepthItem).Price >= price[0] {
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetAsksMaxUpToSumma(target float64) (limit *DepthItem) {
	limit = &DepthItem{}
	summa := 0.0
	d.GetAsks().Ascend(func(i btree.Item) bool {
		summa += i.(*DepthItem).Quantity
		if summa <= target {
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetBidsMaxDownToSumma(target float64) (limit *DepthItem) {
	limit = &DepthItem{}
	summa := 0.0
	d.GetBids().Descend(func(i btree.Item) bool {
		summa += i.(*DepthItem).Quantity
		if summa <= target {
			if limit.Quantity < i.(*DepthItem).Quantity {
				limit.Quantity = i.(*DepthItem).Quantity
				limit.Price = i.(*DepthItem).Price
			}
			return true
		} else {
			return false
		}
	})
	return
}

func (d *Depth) GetSummaOfAsksFromRange(first, last float64, f ...DepthFilter) (askSumma float64) {
	var filter DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*DepthItem) bool { return true }
	}
	d.GetAsks().AscendGreaterOrEqual(&DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*DepthItem)) && i.(*DepthItem).Price <= last {
			askSumma += i.(*DepthItem).Quantity
		}
		return true
	})
	return
}

func (d *Depth) GetSummaOfBidsFromRange(first, last float64, f ...DepthFilter) (bidSumma float64) {
	var filter DepthFilter
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*DepthItem) bool { return true }
	}
	d.GetBids().DescendLessOrEqual(&DepthItem{Price: first}, func(i btree.Item) bool {
		if filter(i.(*DepthItem)) && i.(*DepthItem).Price >= last {
			bidSumma += i.(*DepthItem).Quantity
		}
		return true
	})
	return
}
