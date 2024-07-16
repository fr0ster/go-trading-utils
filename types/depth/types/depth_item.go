package types

import (
	"github.com/google/btree"
)

type (
	DepthItem struct {
		Price    float64
		Quantity float64
	}
	DepthFilter func(*DepthItem) bool
	DepthTester func(result *DepthItem, target *DepthItem) bool
)

func (i *DepthItem) Less(than btree.Item) bool {
	return i.Price < than.(*DepthItem).Price
}

func (i *DepthItem) Equal(than btree.Item) bool {
	return i.Price == than.(*DepthItem).Price
}

// GetAskDeviation implements depth_interface.Depths.
func (d *DepthItem) GetQuantityDeviation(middle float64) float64 {
	return d.Quantity - middle
}
