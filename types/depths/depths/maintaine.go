package depths

import (
	"github.com/google/btree"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
)

// Get implements depth_interface.Depths.
func (d *Depths) GetTree() *btree.BTree {
	return d.tree
}

// Set implements depth_interface.Depths.
func (d *Depths) SetTree(tree *btree.BTree) {
	d.tree = tree
	d.tree.Ascend(func(i btree.Item) bool {
		d.summaQuantity += i.(*items_types.DepthItem).GetQuantity()
		d.summaValue += i.(*items_types.DepthItem).GetValue()
		d.countQuantity++
		return true
	})
}

// Clear implements depth_interface.Depths.
func (d *Depths) Clear() {
	d.tree.Clear(false)
	d.summaQuantity = 0
	d.summaValue = 0
	d.countQuantity = 0
}

// RestrictUp implements depth_interface.Depths.
func (d *Depths) RestrictUp(price items_types.PriceType) {
	prices := make([]items_types.PriceType, 0)
	d.tree.AscendGreaterOrEqual(items_types.New(price), func(i btree.Item) bool {
		prices = append(prices, i.(*items_types.DepthItem).GetPrice())
		return true
	})
	for _, p := range prices {
		d.tree.Delete(items_types.New(p))
	}
}

// RestrictDown implements depth_interface.Depths.
func (d *Depths) RestrictDown(price items_types.PriceType) {
	prices := make([]items_types.PriceType, 0)
	d.tree.DescendLessOrEqual(items_types.New(price), func(i btree.Item) bool {
		prices = append(prices, i.(*items_types.DepthItem).GetPrice())
		return true
	})
	for _, p := range prices {
		d.tree.Delete(items_types.New(p))
	}
}
