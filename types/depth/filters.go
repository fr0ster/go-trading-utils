package depth

// func (d *Depths) getIterator(tree *btree.BTree, summa, max, min *types.QuantityType, f ...types.DepthFilter) func(i btree.Item) bool {
// 	return func(i btree.Item) bool {
// 		var filter types.DepthFilter
// 		pp := i.(*types.DepthItem)
// 		if len(f) > 0 {
// 			filter = f[0]
// 		} else {
// 			filter = func(*types.DepthItem) bool { return true }
// 		}
// 		if filter(pp) {
// 			tree.ReplaceOrInsert(pp)
// 			if summa != nil {
// 				*summa += pp.GetQuantity()
// 			}
// 			if max != nil {
// 				if *max < pp.GetQuantity() {
// 					*max = pp.GetQuantity()
// 				}
// 			}
// 			if min != nil {
// 				if *min > pp.GetQuantity() || *min == 0 {
// 					*min = pp.GetQuantity()
// 				}
// 			}
// 		}
// 		return true // продовжуємо обхід
// 	}
// }

// func (d *Depths) GetFilteredByPercentAsks(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min types.QuantityType) {
// 	tree = btree.New(d.degree)
// 	if len(f) > 0 {
// 		d.AskAscend(d.getIterator(tree, &summa, &max, &min, f[0]))
// 	} else {
// 		d.AskAscend(d.getIterator(tree, &summa, &max, &min))
// 	}
// 	return
// }

// func (d *Depths) GetFilteredByPercentBids(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min types.QuantityType) {
// 	tree = btree.New(d.degree)
// 	if len(f) > 0 {
// 		d.BidDescend(d.getIterator(tree, &summa, &max, &min, f[0]))
// 	} else {
// 		d.BidDescend(d.getIterator(tree, &summa, &max, &min))
// 	}
// 	return
// }

// func (d *Depths) GetSummaOfAsksFromRange(first, last types.PriceType, f ...types.DepthFilter) (askSumma, max, min types.QuantityType) {
// 	var filter types.DepthFilter
// 	if len(f) > 0 {
// 		filter = f[0]
// 	} else {
// 		filter = func(*types.DepthItem) bool { return true }
// 	}
// 	d.GetAsks().DescendRange(types.New(last), types.New(first), func(i btree.Item) bool {
// 		if filter(i.(*types.DepthItem)) {
// 			askSumma += i.(*types.DepthItem).GetQuantity()
// 			if max < i.(*types.DepthItem).GetQuantity() {
// 				max = i.(*types.DepthItem).GetQuantity()
// 			}
// 			if min > i.(*types.DepthItem).GetQuantity() || min == 0 {
// 				min = i.(*types.DepthItem).GetQuantity()
// 			}
// 		}
// 		return true
// 	})
// 	return
// }

// func (d *Depths) GetSummaOfBidsFromRange(first, last types.PriceType, f ...types.DepthFilter) (bidSumma, max, min types.QuantityType) {
// 	var filter types.DepthFilter
// 	if len(f) > 0 {
// 		filter = f[0]
// 	} else {
// 		filter = func(*types.DepthItem) bool { return true }
// 	}
// 	d.GetBids().AscendRange(types.New(last), types.New(first), func(i btree.Item) bool {
// 		if filter(i.(*types.DepthItem)) {
// 			bidSumma += i.(*types.DepthItem).GetQuantity()
// 			if max < i.(*types.DepthItem).GetQuantity() {
// 				max = i.(*types.DepthItem).GetQuantity()
// 			}
// 			if min > i.(*types.DepthItem).GetQuantity() || min == 0 {
// 				min = i.(*types.DepthItem).GetQuantity()
// 			}
// 		}
// 		return true
// 	})
// 	return
// }

func (d *Depths) GetPercentToTarget() float64 {
	return d.percentToTarget
}

// // RestrictAskUp implements depth_interface.Depths.
// func (d *Depths) RestrictAskUp(price types.PriceType) {
// 	prices := make([]types.PriceType, 0)
// 	d.asks.AscendGreaterOrEqual(types.New(price), func(i btree.Item) bool {
// 		prices = append(prices, i.(*types.DepthItem).GetPrice())
// 		return true
// 	})
// 	for _, p := range prices {
// 		d.asks.Delete(types.New(p))
// 	}
// }

// // RestrictBidUp implements depth_interface.Depths.
// func (d *Depths) RestrictBidUp(price types.PriceType) {
// 	prices := make([]types.PriceType, 0)
// 	d.bids.AscendGreaterOrEqual(types.New(price), func(i btree.Item) bool {
// 		prices = append(prices, i.(*types.DepthItem).GetPrice())
// 		return true
// 	})
// 	for _, p := range prices {
// 		d.bids.Delete(types.New(p))
// 	}
// }

// // RestrictAskDown implements depth_interface.Depths.
// func (d *Depths) RestrictAskDown(price types.PriceType) {
// 	prices := make([]types.PriceType, 0)
// 	d.asks.DescendLessOrEqual(types.New(price), func(i btree.Item) bool {
// 		prices = append(prices, i.(*types.DepthItem).GetPrice())
// 		return true
// 	})
// 	for _, p := range prices {
// 		d.asks.Delete(types.New(p))
// 	}
// }

// // RestrictBidDown implements depth_interface.Depths.
// func (d *Depths) RestrictBidDown(price types.PriceType) {
// 	prices := make([]types.PriceType, 0)
// 	d.bids.DescendLessOrEqual(types.New(price), func(i btree.Item) bool {
// 		prices = append(prices, i.(*types.DepthItem).GetPrice())
// 		return true
// 	})
// 	for _, p := range prices {
// 		d.bids.Delete(types.New(p))
// 	}
// }
