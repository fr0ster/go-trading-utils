package types_test

import (
	"testing"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/stretchr/testify/assert"
)

func TestGetNormalizedPrice(t *testing.T) {
	price := types.NewNormalizedItem(150, 3, 2, false).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(100.0), price)
	price = types.NewNormalizedItem(1.552, 3, 2, false).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(1.5), price)
	price = types.NewNormalizedItem(1.941, 3, 2, false).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(1.9), price)
	price = types.NewNormalizedItem(80, 3, 2, false).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(80.0), price)
	price = types.NewNormalizedItem(45.5, 3, 2, false).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(45.0), price)
	price = types.NewNormalizedItem(150, 3, 2, true).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(200.0), price)
	price = types.NewNormalizedItem(1.552, 3, 2, true).GetNormalizedPrice()
	assert.Equal(t, types.PriceType(1.6), price)
	price = types.NewNormalizedItem(1.941, 3, 2, true).GetNormalizedPrice() // GetNormalizedPrice
	assert.Equal(t, types.PriceType(2.0), price)
}
