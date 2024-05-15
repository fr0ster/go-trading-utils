package grid_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/types/grid"
)

func TestRecord(t *testing.T) {
	// Create a new grid
	g := grid.NewLine(1, 10.5, 12.5, 8.5)

	// Test GetOrderId
	if g.GetOrderId() != 1 {
		t.Errorf("Expected GetOrderId to return 1, but got %d", g.GetOrderId())
	}

	// Test GetPrice
	if g.GetPrice() != 10.5 {
		t.Errorf("Expected GetPrice to return 10.5, but got %f", g.GetPrice())
	}

	// Test GetUpPrice
	if g.GetUpPrice() != 12.5 {
		t.Errorf("Expected GetUpPrice to return 12.5, but got %f", g.GetUpPrice())
	}

	// Test GetDownPrice
	if g.GetDownPrice() != 8.5 {
		t.Errorf("Expected GetDownPrice to return 8.5, but got %f", g.GetDownPrice())
	}

	// Test SetOrderId
	g.SetOrderId(4)
	if g.GetOrderId() != 4 {
		t.Errorf("Expected GetOrderId to return 4 after SetOrderId, but got %d", g.GetOrderId())
	}

	// Test SetPrice
	g.SetPrice(15.5)
	if g.GetPrice() != 15.5 {
		t.Errorf("Expected GetPrice to return 15.5 after SetPrice, but got %f", g.GetPrice())
	}

	// Test SetUpPrice
	g.SetUpPrice(18.5)
	if g.GetUpPrice() != 18.5 {
		t.Errorf("Expected GetUpPrice to return 18.5 after SetUpPrice, but got %f", g.GetUpPrice())
	}

	// Test SetDownPrice
	g.SetDownPrice(9.5)
	if g.GetDownPrice() != 9.5 {
		t.Errorf("Expected GetDownPrice to return 9.5 after SetDownPrice, but got %f", g.GetDownPrice())
	}

	// Test Equals
	other := grid.NewLine(1, 15.5, 12.5, 8.5)
	if !g.Equals(other) {
		t.Errorf("Expected Equals to return true for two identical grids, but got false")
	}

	// Test Less
	other = grid.NewLine(4, 20.5, 18.5, 9.5)
	if !g.Less(other) {
		t.Errorf("Expected Less to return true for g < other, but got false")
	}
}
