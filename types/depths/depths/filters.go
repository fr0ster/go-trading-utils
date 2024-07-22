package depths

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/google/btree"
)

func (d *Depths) GetFiltered(up UpOrDown, filter ...items_types.DepthFilter) (tree *Depths) {
	getIterator := func(tree *Depths, f ...items_types.DepthFilter) func(i btree.Item) bool {
		return func(i btree.Item) bool {
			var filter items_types.DepthFilter
			pp := i.(*items_types.DepthItem)
			if len(f) > 0 {
				filter = f[0]
			} else {
				filter = func(*items_types.DepthItem) bool { return true }
			}
			if filter(pp) {
				tree.Set(pp)
			}
			return true // продовжуємо обхід
		}
	}
	tree = New(d.degree, d.symbol)
	if up {
		d.GetTree().Ascend(getIterator(tree, filter...))
	} else {
		d.GetTree().Descend(getIterator(tree, filter...))
	}
	return
}
