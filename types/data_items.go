package types

import "github.com/google/btree"

type (
	DepthItemType struct {
		Price    float64
		Quantity float64
	}
)

// DepthItemType - тип для зберігання заявок в стакані
func (i DepthItemType) Less(than btree.Item) bool {
	return i.Price < than.(DepthItemType).Price
}

func (i DepthItemType) Equal(than btree.Item) bool {
	return i.Price == than.(DepthItemType).Price
}
