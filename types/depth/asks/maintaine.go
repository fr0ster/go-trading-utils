package asks

import (
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/google/btree"
)

// Get implements depth_interface.Depths.
func (d *Asks) GetTree() *btree.BTree {
	return d.tree.GetTree()
}

// Set implements depth_interface.Depths.
func (d *Asks) SetTree(tree *btree.BTree) {
	d.tree.SetTree(tree)
}

// Clear implements depth_interface.Depths.
func (d *Asks) Clear() {
	d.tree.Clear()
}

// RestrictUp implements depth_interface.Depths.
func (d *Asks) RestrictUp(price items_types.PriceType) {
	d.tree.RestrictUp(price)
}

// RestrictDown implements depth_interface.Depths.
func (d *Asks) RestrictDown(price items_types.PriceType) {
	d.tree.RestrictDown(price)
}
