package grid_test

import (
	"testing"

	types "github.com/fr0ster/go-trading-utils/types"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"

	"github.com/stretchr/testify/assert"
)

func TestGridOverTree(t *testing.T) {
	// Create a new tree
	tree := grid_types.New()

	// Insert the grid into the tree
	tree.Set(grid_types.NewRecord(13, 30.5, 5, 0.0, 25.5, types.SideTypeSell))
	tree.Set(grid_types.NewRecord(7, 25.5, 5, 30.5, 20.5, types.SideTypeSell))
	tree.Set(grid_types.NewRecord(0, 20.5, 5, 25.5, 15.5, types.SideTypeNone))
	tree.Set(grid_types.NewRecord(4, 15.5, 5, 20.5, 10.5, types.SideTypeBuy))
	tree.Set(grid_types.NewRecord(11, 10.5, 5, 15.5, 0, types.SideTypeBuy))

	// Test Get
	raw := tree.Get(&grid_types.Record{Price: 10.5})
	if raw != nil {
		assert.Equal(t, int64(11), raw.(*grid_types.Record).GetOrderId())
		assert.Equal(t, items_types.PriceType(10.5), raw.(*grid_types.Record).GetPrice())
		assert.Equal(t, items_types.QuantityType(5.0), raw.(*grid_types.Record).GetQuantity())
		assert.Equal(t, items_types.PriceType(15.5), raw.(*grid_types.Record).GetUpPrice())
		assert.Equal(t, items_types.PriceType(0.0), raw.(*grid_types.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeBuy, raw.(*grid_types.Record).GetOrderSide())
	}
	raw = tree.Get(&grid_types.Record{Price: 20.5})
	if raw != nil {
		assert.Equal(t, int64(0), raw.(*grid_types.Record).GetOrderId())
		assert.Equal(t, items_types.PriceType(20.5), raw.(*grid_types.Record).GetPrice())
		assert.Equal(t, items_types.QuantityType(5.0), raw.(*grid_types.Record).GetQuantity())
		assert.Equal(t, items_types.PriceType(25.5), raw.(*grid_types.Record).GetUpPrice())
		assert.Equal(t, items_types.PriceType(15.5), raw.(*grid_types.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeNone, raw.(*grid_types.Record).GetOrderSide())
	}
	raw = tree.Get(&grid_types.Record{Price: 30.5})
	if raw != nil {
		assert.Equal(t, int64(13), raw.(*grid_types.Record).GetOrderId())
		assert.Equal(t, items_types.PriceType(30.5), raw.(*grid_types.Record).GetPrice())
		assert.Equal(t, items_types.QuantityType(5.0), raw.(*grid_types.Record).GetQuantity())
		assert.Equal(t, items_types.PriceType(0.0), raw.(*grid_types.Record).GetUpPrice())
		assert.Equal(t, items_types.PriceType(25.5), raw.(*grid_types.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeSell, raw.(*grid_types.Record).GetOrderSide())
	}
	raw = tree.Get(&grid_types.Record{Price: 25.5})
	if raw != nil {
		assert.Equal(t, int64(7), raw.(*grid_types.Record).GetOrderId())
		assert.Equal(t, items_types.PriceType(25.5), raw.(*grid_types.Record).GetPrice())
		assert.Equal(t, items_types.QuantityType(5.0), raw.(*grid_types.Record).GetQuantity())
		assert.Equal(t, items_types.PriceType(30.5), raw.(*grid_types.Record).GetUpPrice())
		assert.Equal(t, items_types.PriceType(20.5), raw.(*grid_types.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeSell, raw.(*grid_types.Record).GetOrderSide())
	}
	raw = tree.Get(&grid_types.Record{Price: 15.5})
	if raw != nil {
		assert.Equal(t, int64(4), raw.(*grid_types.Record).GetOrderId())
		assert.Equal(t, items_types.PriceType(15.5), raw.(*grid_types.Record).GetPrice())
		assert.Equal(t, items_types.QuantityType(5.0), raw.(*grid_types.Record).GetQuantity())
		assert.Equal(t, items_types.PriceType(20.5), raw.(*grid_types.Record).GetUpPrice())
		assert.Equal(t, items_types.PriceType(10.5), raw.(*grid_types.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeBuy, raw.(*grid_types.Record).GetOrderSide())
	}
}
