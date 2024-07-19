package depths

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

func (d *Depths) getIterator(tree *btree.BTree, summa, max, min *items_types.QuantityType, f ...items_types.DepthFilter) func(i btree.Item) bool {
	return func(i btree.Item) bool {
		var filter items_types.DepthFilter
		pp := i.(*items_types.DepthItem)
		if len(f) > 0 {
			filter = f[0]
		} else {
			filter = func(*items_types.DepthItem) bool { return true }
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

func (d *Depths) GetFilteredByPercent(f ...items_types.DepthFilter) (tree *btree.BTree, summa, max, min items_types.QuantityType) {
	tree = btree.New(d.degree)
	if len(f) > 0 {
		d.GetTree().Ascend(d.getIterator(tree, &summa, &max, &min, f[0]))
	} else {
		d.GetTree().Ascend(d.getIterator(tree, &summa, &max, &min))
	}
	return
}

func (d *Depths) GetSummaByRange(first, last items_types.PriceType, f ...items_types.DepthFilter) (summa, max, min items_types.QuantityType) {
	var (
		filter items_types.DepthFilter
		ranger func(lessOrEqual, greaterThan btree.Item, iterator btree.ItemIterator)
	)
	if len(f) > 0 {
		filter = f[0]
	} else {
		filter = func(*items_types.DepthItem) bool { return true }
	}
	if first <= last {
		ranger = d.GetTree().DescendRange
	} else {
		ranger = d.GetTree().AscendRange
	}
	ranger(items_types.New(last), items_types.New(first), func(i btree.Item) bool {
		if filter(i.(*items_types.DepthItem)) {
			summa += i.(*items_types.DepthItem).GetQuantity()
			if max < i.(*items_types.DepthItem).GetQuantity() {
				max = i.(*items_types.DepthItem).GetQuantity()
			}
			if min > i.(*items_types.DepthItem).GetQuantity() || min == 0 {
				min = i.(*items_types.DepthItem).GetQuantity()
			}
		}
		return true
	})
	return
}
