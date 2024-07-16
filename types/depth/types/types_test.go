package types_test

import (
	"testing"

	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	"github.com/stretchr/testify/assert"
)

func TestGetNormalizedPrice(t *testing.T) {
	price, err := types.NewNormalizedItem(150, 3, 2, false).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 100.0, price)
	price, err = types.NewNormalizedItem(1.552, 3, 2, false).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 1.5, price)
	price, err = types.NewNormalizedItem(1.941, 3, 2, false).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 1.9, price)
	price, err = types.NewNormalizedItem(80, 3, 2, false).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 80.0, price)
	price, err = types.NewNormalizedItem(45.5, 3, 2, false).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 45.0, price)
	price, err = types.NewNormalizedItem(150, 3, 2, true).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 200.0, price)
	price, err = types.NewNormalizedItem(1.552, 3, 2, true).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 1.6, price)
	price, err = types.NewNormalizedItem(1.941, 3, 2, true).GetNormalizedPrice()
	assert.Nil(t, err)
	assert.Equal(t, 2.0, price)
}
