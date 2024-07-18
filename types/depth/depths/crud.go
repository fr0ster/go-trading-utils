package depths

import (
	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Get implements depth_interface.Depths.
func (a *Depths) Get(item *types.DepthItem) *types.DepthItem {
	if val := a.tree.Get(item); val != nil {
		return val.(*types.DepthItem)
	} else {
		return nil
	}
}

// Set implements depth_interface.Depths.
func (d *Depths) Set(item *types.DepthItem) (err error) {
	if old := d.tree.Get(item); old != nil {
		d.summaQuantity += item.GetQuantity() - old.(*types.DepthItem).GetQuantity()
	} else {
		d.summaQuantity += item.GetQuantity()
		d.countQuantity++
	}
	d.tree.ReplaceOrInsert(item)
	return
}

// Delete implements depth_interface.Depths.
func (d *Depths) Delete(item *types.DepthItem) {
	old := d.tree.Get(item)
	if old != nil {
		d.summaQuantity -= old.(*types.DepthItem).GetQuantity()
		d.countQuantity--
		d.tree.Delete(item)
	}
}

// Update implements depth_interface.Depths.
func (d *Depths) Update(item *types.DepthItem) bool {
	if item.GetQuantity() == 0 {
		d.Delete(item)
	} else {
		d.Set(item)
	}
	return true
}

// Count implements depth_interface.Depths.
func (d *Depths) Count() int {
	return d.countQuantity
}

// Symbol implements depth_interface.Depths.
func (d *Depths) Symbol() string {
	return d.symbol
}

func (d *Depths) GetSummaQuantity() types.QuantityType {
	return d.summaQuantity
}
