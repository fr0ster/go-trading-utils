package depths

import (
	"math"

	"github.com/google/btree"

	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// SetItems implements depth_interface.Depths.
func (d *Depths) SetItems(asks *btree.BTree) (summaQuantity types.QuantityType, countQuantity int) {
	d.tree.Clear(false)
	asks.Ascend(func(i btree.Item) bool {
		summaQuantity += i.(*types.DepthItem).GetQuantity()
		countQuantity++
		d.Set(i.(*types.DepthItem))
		return true
	})
	return
}

// Clear implements depth_interface.Depths.
func (d *Depths) Clear() {
	d.tree.Clear(false)
	d.summaQuantity = 0
	d.countQuantity = 0
}

func (d *Depths) GetMiddleQuantity() types.QuantityType {
	return d.summaQuantity / types.QuantityType(d.countQuantity)
}

func (d *Depths) GetStandardDeviation() float64 {
	summaSquares := 0.0
	d.GetTree().Ascend(func(i btree.Item) bool {
		depth := i.(*types.DepthItem)
		summaSquares += depth.GetQuantityDeviation(d.GetMiddleQuantity()) * depth.GetQuantityDeviation(d.GetMiddleQuantity())
		return true
	})
	return math.Sqrt(summaSquares / float64(d.Count()))
}
