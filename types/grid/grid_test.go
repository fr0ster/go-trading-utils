package grid_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/types/grid"
	"github.com/stretchr/testify/assert"
)

func TestGridOverTree(t *testing.T) {
	// Create a new tree
	tree := grid.New()

	// Insert the grid into the tree
	tree.Set(grid.NewRecord(1, 10.5, 15.5, 0))
	tree.Set(grid.NewRecord(4, 15.5, 20.5, 10.5))
	tree.Set(grid.NewRecord(7, 20.5, 25.5, 15.5))
	tree.Set(grid.NewRecord(10, 25.5, 30.5, 20.5))
	tree.Set(grid.NewRecord(13, 30.5, 35.5, 25.5))
	tree.Set(grid.NewRecord(16, 35.5, 0, 30.5))

	// Test Get
	raw := tree.Get(&grid.Record{Price: 10.5})
	if raw != nil {
		assert.Equal(t, 1, raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 10.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 15.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 0.0, raw.(*grid.Record).GetDownPrice())
	}
	raw = tree.Get(&grid.Record{OrderId: 1})
	if raw != nil {
		assert.Equal(t, 1, raw.(*grid.Record).GetOrderId())
		assert.Equal(t, 10.5, raw.(*grid.Record).GetPrice())
		assert.Equal(t, 15.5, raw.(*grid.Record).GetUpPrice())
		assert.Equal(t, 0.0, raw.(*grid.Record).GetDownPrice())
	}
}
