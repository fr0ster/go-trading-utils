package filters_test

import (
	"testing"

	"github.com/fr0ster/go-binance-utils/spot/info/symbols/filters"
	"github.com/stretchr/testify/assert"
)

func TestFilters_GetFilter(t *testing.T) {
	// Create a new Filters instance
	f := filters.NewFilters(10)

	// Insert some filters
	f.Insert("filter1", map[string]interface{}{"param1": "value1"})
	f.Insert("filter2", map[string]interface{}{"param2": "value2"})

	// Get a filter by symbol
	filter := f.GetFilter("filter1")

	// Assert that the filter is not nil
	assert.NotNil(t, filter)

	// Assert that the filter has the expected values
	assert.Equal(t, filters.FilterName("filter1"), filter.FilterName)
	assert.Equal(t, filters.Filter(filters.Filter{"param1": "value1"}), filter.Filter)
}

func TestFilters_DeleteFilter(t *testing.T) {
	// Create a new Filters instance
	f := filters.NewFilters(10)

	// Insert a filter
	f.Insert("filter1", map[string]interface{}{"param1": "value1"})

	// Delete the filter
	f.DeleteFilter("filter1")

	// Get the filter by symbol
	filter := f.GetFilter("filter1")

	// Assert that the filter is nil
	assert.Nil(t, filter)
}

// Add more test cases for other functions as needed
