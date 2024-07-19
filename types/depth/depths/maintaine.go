package depths

import (
	"math"

	"github.com/google/btree"

	types "github.com/fr0ster/go-trading-utils/types/depth/items"
)

// Clear implements depth_interface.Depths.
func (d *Depths) Clear() {
	d.tree.Clear(false)
	d.summaQuantity = 0
	d.summaValue = 0
	d.countQuantity = 0
}

func (d *Depths) GetMiddleQuantity() types.QuantityType {
	return d.summaQuantity / types.QuantityType(d.countQuantity)
}

func (d *Depths) GetMiddleValue() types.ValueType {
	return d.summaValue / types.ValueType(d.countQuantity)
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
