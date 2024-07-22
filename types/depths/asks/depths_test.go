package asks_test

import (
	"testing"

	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestAsksGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	asks := asks_types.New(degree, "BTCUSDT")
	asks.Set(items_types.NewAsk(100, 10))
	asks.Set(items_types.NewAsk(200, 20))
	asks.Set(items_types.NewAsk(300, 30))
	asks.Set(items_types.NewAsk(400, 40))
	asks.Set(items_types.NewAsk(500, 50))

	assert.Equal(t, 5, asks.Count())
	assert.Equal(t, items_types.QuantityType(150), asks.GetSummaQuantity())
	assert.Equal(t, items_types.PriceType(100), (asks.Get(items_types.NewAsk(100))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(0), (asks.Get(items_types.NewAsk(600))).GetDepthItem().GetPrice())

	assert.Equal(t, items_types.QuantityType(10), asks.Get(items_types.NewAsk(100)).GetDepthItem().GetQuantity())
	asks.Get(items_types.NewAsk(100)).GetDepthItem().SetQuantity(200)
	assert.Equal(t, items_types.PriceType(100), asks.Get(items_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.QuantityType(200), asks.Get(items_types.NewAsk(100)).GetDepthItem().GetQuantity())

	item := asks.Get(items_types.NewAsk(100))
	item.GetDepthItem().SetPrice(600)
	asks.Delete(item)
	asks.Set(items_types.NewAsk(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	assert.Equal(t, items_types.PriceType(0), (asks.Get(items_types.NewAsk(100))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(600), (asks.Get(items_types.NewAsk(600))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.QuantityType(200), (asks.Get(items_types.NewAsk(600))).GetDepthItem().GetQuantity())
}

func TestGetAndSetAsks(t *testing.T) {
	// TODO: Add test cases.
	depth := asks_types.New(degree, "BTCUSDT")
	depth.Set(items_types.NewAsk(100, 10))
	depth.Set(items_types.NewAsk(200, 20))
	depth.Set(items_types.NewAsk(300, 30))
	depth.Set(items_types.NewAsk(400, 40))
	depth.Set(items_types.NewAsk(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, items_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, items_types.PriceType(100), depth.Get(items_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(0), (depth.Get(items_types.NewAsk(600))).GetDepthItem().GetPrice())

	otherDepth := asks_types.New(degree, "BTCUSDT")
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, items_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, items_types.PriceType(100), depth.Get(items_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(0), (depth.Get(items_types.NewAsk(600))).GetDepthItem().GetPrice())
}

func TestGetMaxQuantity(t *testing.T) {
	// TODO: Add test cases.
	depth := asks_types.New(degree, "BTCUSDT")
	depth.Set(items_types.NewAsk(100, 10))
	depth.Set(items_types.NewAsk(200, 20))
	depth.Set(items_types.NewAsk(300, 30))
	depth.Set(items_types.NewAsk(400, 20))
	depth.Set(items_types.NewAsk(500, 10))

	min, max := depth.GetMinMaxByQuantity()
	assert.Equal(t, items_types.QuantityType(30), max.GetQuantity())
	assert.Equal(t, items_types.PriceType(300), max.GetPrice())
	assert.Equal(t, items_types.QuantityType(10), min.GetQuantity())
	assert.Equal(t, items_types.PriceType(100), min.GetPrice())
}
