package asks_test

import (
	"testing"

	asks_types "github.com/fr0ster/go-trading-utils/types/depth/asks"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	item_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestAsksGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	asks := asks_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	asks.Set(item_types.NewAsk(100, 10))
	asks.Set(item_types.NewAsk(200, 20))
	asks.Set(item_types.NewAsk(300, 30))
	asks.Set(item_types.NewAsk(400, 40))
	asks.Set(item_types.NewAsk(500, 50))

	assert.Equal(t, 5, asks.Count())
	assert.Equal(t, item_types.QuantityType(150), asks.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), (asks.Get(item_types.NewAsk(100))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (asks.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())

	assert.Equal(t, item_types.QuantityType(10), asks.Get(item_types.NewAsk(100)).GetDepthItem().GetQuantity())
	asks.Get(item_types.NewAsk(100)).GetDepthItem().SetQuantity(200)
	assert.Equal(t, item_types.PriceType(100), asks.Get(item_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.QuantityType(200), asks.Get(item_types.NewAsk(100)).GetDepthItem().GetQuantity())

	item := asks.Get(item_types.NewAsk(100))
	item.GetDepthItem().SetPrice(600)
	asks.Delete(item)
	asks.Set(item_types.NewAsk(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	assert.Equal(t, item_types.PriceType(0), (asks.Get(item_types.NewAsk(100))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(600), (asks.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.QuantityType(200), (asks.Get(item_types.NewAsk(600))).GetDepthItem().GetQuantity())
}

func TestGetAndSetAsks(t *testing.T) {
	// TODO: Add test cases.
	depth := asks_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.NewAsk(100, 10))
	depth.Set(item_types.NewAsk(200, 20))
	depth.Set(item_types.NewAsk(300, 30))
	depth.Set(item_types.NewAsk(400, 40))
	depth.Set(item_types.NewAsk(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())

	otherDepth := asks_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())
}

func TestGetMaxQuantity(t *testing.T) {
	// TODO: Add test cases.
	depth := asks_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.NewAsk(100, 10))
	depth.Set(item_types.NewAsk(200, 20))
	depth.Set(item_types.NewAsk(300, 30))
	depth.Set(item_types.NewAsk(400, 20))
	depth.Set(item_types.NewAsk(500, 10))

	min, max := depth.GetMinMaxQuantity()
	assert.Equal(t, item_types.QuantityType(30), max.GetQuantity())
	assert.Equal(t, item_types.PriceType(300), max.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), min.GetQuantity())
	assert.Equal(t, item_types.PriceType(100), min.GetPrice())
}
