package depths_test

import (
	"testing"

	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	item_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestDepthsGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10))
	depth.Set(item_types.New(200, 20))
	depth.Set(item_types.New(300, 30))
	depth.Set(item_types.New(400, 40))
	depth.Set(item_types.New(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())

	assert.Equal(t, item_types.QuantityType(10), depth.Get(item_types.New(100)).GetQuantity())
	depth.Get(item_types.New(100)).SetQuantity(200)
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.QuantityType(200), depth.Get(item_types.New(100)).GetQuantity())

	item := depth.Get(item_types.New(100))
	item.SetPrice(600)
	depth.Delete(item)
	depth.Set(item_types.New(item.GetPrice(), item.GetQuantity()))
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(100))).GetPrice())
	assert.Equal(t, item_types.PriceType(600), (depth.Get(item_types.New(600))).GetPrice())
	assert.Equal(t, item_types.QuantityType(200), (depth.Get(item_types.New(600))).GetQuantity())
}

func TestAsksGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	asks := depths_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	asks.Set(item_types.NewAsk(100, 10))
	asks.Set(item_types.NewAsk(200, 20))
	asks.Set(item_types.NewAsk(300, 30))
	asks.Set(item_types.NewAsk(400, 40))
	asks.Set(item_types.NewAsk(500, 50))

	assert.Equal(t, 5, asks.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), asks.GetDepths().GetSummaQuantity())
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

func TestBidsGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	bids := depths_types.NewBids(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	bids.Set(item_types.NewBid(100, 10))
	bids.Set(item_types.NewBid(200, 20))
	bids.Set(item_types.NewBid(300, 30))
	bids.Set(item_types.NewBid(400, 40))
	bids.Set(item_types.NewBid(500, 50))

	assert.Equal(t, 5, bids.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), bids.GetDepths().GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), (bids.Get(item_types.NewBid(100))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (bids.Get(item_types.NewBid(600))).GetDepthItem().GetPrice())

	assert.Equal(t, item_types.QuantityType(10), bids.Get(item_types.NewBid(100)).GetDepthItem().GetQuantity())
	bids.Get(item_types.NewBid(100)).GetDepthItem().SetQuantity(200)
	assert.Equal(t, item_types.PriceType(100), bids.Get(item_types.NewBid(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.QuantityType(200), bids.Get(item_types.NewBid(100)).GetDepthItem().GetQuantity())

	item := bids.Get(item_types.NewBid(100))
	item.GetDepthItem().SetPrice(600)
	bids.Delete(item)
	bids.Set(item_types.NewBid(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	assert.Equal(t, item_types.PriceType(0), (bids.Get(item_types.NewBid(100))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(600), (bids.Get(item_types.NewBid(600))).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.QuantityType(200), (bids.Get(item_types.NewBid(600))).GetDepthItem().GetQuantity())
}

func TestGetAndSetDepths(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10))
	depth.Set(item_types.New(200, 20))
	depth.Set(item_types.New(300, 30))
	depth.Set(item_types.New(400, 40))
	depth.Set(item_types.New(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())

	otherDepth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())
}

func TestGetAndSetAsks(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.NewAsk(100, 10))
	depth.Set(item_types.NewAsk(200, 20))
	depth.Set(item_types.NewAsk(300, 30))
	depth.Set(item_types.NewAsk(400, 40))
	depth.Set(item_types.NewAsk(500, 50))

	assert.Equal(t, 5, depth.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetDepths().GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())

	otherDepth := depths_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetDepths().GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewAsk(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewAsk(600))).GetDepthItem().GetPrice())
}

func TestGetAndSetBids(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.NewBids(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.NewBid(100, 10))
	depth.Set(item_types.NewBid(200, 20))
	depth.Set(item_types.NewBid(300, 30))
	depth.Set(item_types.NewBid(400, 40))
	depth.Set(item_types.NewBid(500, 50))

	assert.Equal(t, 5, depth.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetDepths().GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewBid(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewBid(600))).GetDepthItem().GetPrice())

	otherDepth := depths_types.NewAsks(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.GetDepths().Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetDepths().GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.NewBid(100)).GetDepthItem().GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.NewBid(600))).GetDepthItem().GetPrice())
}
