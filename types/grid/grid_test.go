package grid_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/types/grid"
	"github.com/stretchr/testify/assert"
)

func TestGridOverTree(t *testing.T) {
	// Create a new tree
	tree := grid.New()

	// Insert the grid into the tree
	tree.Set(grid.NewRecord(13, 30.5, 0.0, 25.5, types.SideTypeSell))
	tree.Set(grid.NewRecord(7, 25.5, 30.5, 20.5, types.SideTypeSell))
	tree.Set(grid.NewRecord(0, 20.5, 25.5, 15.5, types.SideTypeNone))
	tree.Set(grid.NewRecord(4, 15.5, 20.5, 10.5, types.SideTypeBuy))
	tree.Set(grid.NewRecord(11, 10.5, 15.5, 0, types.SideTypeBuy))

	// Test Get
	raw := tree.Get(&grid.Record{Price: 10.5})
	if raw != nil {
		assert.Equal(t, int64(11), raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 10.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 15.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 0.0, raw.(*grid.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeBuy, raw.(*grid.Record).GetOrderSide())
	}
	raw = tree.Get(&grid.Record{Price: 20.5})
	if raw != nil {
		assert.Equal(t, int64(0), raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 20.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 25.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 15.5, raw.(*grid.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeNone, raw.(*grid.Record).GetOrderSide())
	}
	raw = tree.Get(&grid.Record{Price: 30.5})
	if raw != nil {
		assert.Equal(t, int64(13), raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 30.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 0.0, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 25.5, raw.(*grid.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeSell, raw.(*grid.Record).GetOrderSide())
	}
	raw = tree.Get(&grid.Record{Price: 25.5})
	if raw != nil {
		assert.Equal(t, int64(7), raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 25.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 30.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 20.5, raw.(*grid.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeSell, raw.(*grid.Record).GetOrderSide())
	}
	raw = tree.Get(&grid.Record{Price: 15.5})
	if raw != nil {
		assert.Equal(t, int64(4), raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 15.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 20.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 10.5, raw.(*grid.Record).GetDownPrice())
		assert.Equal(t, types.SideTypeBuy, raw.(*grid.Record).GetOrderSide())
	}
}
