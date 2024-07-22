package bids

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/google/btree"
)

// Get implements depth_interface.Depths.
func (d *Bids) GetTree() *btree.BTree {
	return d.tree.GetTree()
}

// Set implements depth_interface.Depths.
func (d *Bids) SetTree(tree *btree.BTree) {
	d.tree.SetTree(tree)
}

// Clear implements depth_interface.Depths.
func (d *Bids) Clear() {
	d.tree.Clear()
}

// RestrictUp implements depth_interface.Depths.
func (d *Bids) RestrictUp(price items_types.PriceType) {
	d.tree.RestrictUp(price)
}

// RestrictDown implements depth_interface.Depths.
func (d *Bids) RestrictDown(price items_types.PriceType) {
	d.tree.RestrictDown(price)
}
