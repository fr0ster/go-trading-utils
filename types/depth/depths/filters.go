package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Depths) getIterator(tree *btree.BTree, summa, max, min *types.QuantityType, f ...types.DepthFilter) func(i btree.Item) bool {
	return func(i btree.Item) bool {
		var filter types.DepthFilter
		pp := i.(*types.DepthItem)
		if len(f) > 0 {
			filter = f[0]
		} else {
			filter = func(*types.DepthItem) bool { return true }
		}
		if filter(pp) {
			tree.ReplaceOrInsert(pp)
			if summa != nil {
				*summa += pp.GetQuantity()
			}
			if max != nil {
				if *max < pp.GetQuantity() {
					*max = pp.GetQuantity()
				}
			}
			if min != nil {
				if *min > pp.GetQuantity() || *min == 0 {
					*min = pp.GetQuantity()
				}
			}
		}
		return true // продовжуємо обхід
	}
}

func (d *Depths) GetFilteredByPercent(f ...types.DepthFilter) (tree *btree.BTree, summa, max, min types.QuantityType) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.GetTree().Ascend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.GetTree().Ascend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depths) GetSummaByRange(first, last types.PriceType, f ...types.DepthFilter) (summa, max, min types.QuantityType) {
	var (
		filter types.DepthFilter
		ranger func(lessOrEqual, greaterThan btree.Item, iterator btree.ItemIterator)
	)
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*types.DepthItem) bool { return true }
	}
	if first <= last {
		ranger = d.GetTree().DescendRange
	} else {
		ranger = d.GetTree().AscendRange
	}
	ranger(types.New(last), types.New(first), func(i btree.Item) bool {
		if filter(i.(*types.DepthItem)) {
			summa += i.(*types.DepthItem).GetQuantity()
			if max < i.(*types.DepthItem).GetQuantity() {
				max = i.(*types.DepthItem).GetQuantity()
			}
			if min > i.(*types.DepthItem).GetQuantity() || min == 0 {
				min = i.(*types.DepthItem).GetQuantity()
			}
		}
		return true
	})
	return
}

// RestrictUp implements depth_interface.Depths.
func (d *Depths) RestrictUp(price types.PriceType) {
	prices := make([]types.PriceType, 0)
	d.tree.AscendGreaterOrEqual(types.New(price), func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).GetPrice())
		return true
	})
	for _, p := range prices {
		d.tree.Delete(types.New(p))
	}
}

// RestrictDown implements depth_interface.Depths.
func (d *Depths) RestrictDown(price types.PriceType) {
	prices := make([]types.PriceType, 0)
	d.tree.DescendLessOrEqual(types.New(price), func(i btree.Item) bool {
		prices = append(prices, i.(*types.DepthItem).GetPrice())
		return true
	})
	for _, p := range prices {
		d.tree.Delete(types.New(p))
	}
}
