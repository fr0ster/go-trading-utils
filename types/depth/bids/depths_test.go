package bids_test

import (
	"testing"

	bids_types "github.com/fr0ster/go-trading-utils/types/depth/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestBidsGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	bids := bids_types.New(degree, "BTCUSDT")
	bids.Set(items_types.NewBid(100, 10))
	bids.Set(items_types.NewBid(200, 20))
	bids.Set(items_types.NewBid(300, 30))
	bids.Set(items_types.NewBid(400, 40))
	bids.Set(items_types.NewBid(500, 50))

	assert.Equal(t, 5, bids.Count())
	assert.Equal(t, items_types.QuantityType(150), bids.GetSummaQuantity())
	assert.Equal(t, items_types.PriceType(100), (bids.Get(items_types.NewBid(100))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(0), (bids.Get(items_types.NewBid(600))).GetDepthItem().GetPrice())

	assert.Equal(t, items_types.QuantityType(10), bids.Get(items_types.NewBid(100)).GetDepthItem().GetQuantity())
	bids.Get(items_types.NewBid(100)).GetDepthItem().SetQuantity(200)
	assert.Equal(t, items_types.PriceType(100), bids.Get(items_types.NewBid(100)).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.QuantityType(200), bids.Get(items_types.NewBid(100)).GetDepthItem().GetQuantity())

	item := bids.Get(items_types.NewBid(100))
	item.GetDepthItem().SetPrice(600)
	bids.Delete(item)
	bids.Set(items_types.NewBid(item.GetDepthItem().GetPrice(), item.GetDepthItem().GetQuantity()))
	assert.Equal(t, items_types.PriceType(0), (bids.Get(items_types.NewBid(100))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(600), (bids.Get(items_types.NewBid(600))).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.QuantityType(200), (bids.Get(items_types.NewBid(600))).GetDepthItem().GetQuantity())
}

func TestGetAndSetBids(t *testing.T) {
	// TODO: Add test cases.
	depth := bids_types.New(degree, "BTCUSDT")
	depth.Set(items_types.NewBid(100, 10))
	depth.Set(items_types.NewBid(200, 20))
	depth.Set(items_types.NewBid(300, 30))
	depth.Set(items_types.NewBid(400, 40))
	depth.Set(items_types.NewBid(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, items_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, items_types.PriceType(100), depth.Get(items_types.NewBid(100)).GetDepthItem().GetPrice())
	assert.Equal(t, items_types.PriceType(0), (depth.Get(items_types.NewBid(600))).GetDepthItem().GetPrice())
}

func TestGetMaxQuantity(t *testing.T) {
	// TODO: Add test cases.
	depth := bids_types.New(degree, "BTCUSDT")
	depth.Set(items_types.NewBid(100, 10))
	depth.Set(items_types.NewBid(200, 20))
	depth.Set(items_types.NewBid(300, 30))
	depth.Set(items_types.NewBid(400, 20))
	depth.Set(items_types.NewBid(500, 10))

	min, max := depth.GetMinMaxByQuantity()
	assert.Equal(t, items_types.QuantityType(30), max.GetQuantity())
	assert.Equal(t, items_types.PriceType(300), max.GetPrice())
	assert.Equal(t, items_types.QuantityType(10), min.GetQuantity())
	assert.Equal(t, items_types.PriceType(500), min.GetPrice())

	min, max = depth.GetMinMaxByQuantity()
	assert.Equal(t, items_types.QuantityType(30), max.GetQuantity())
	assert.Equal(t, items_types.PriceType(300), max.GetPrice())
	assert.Equal(t, items_types.QuantityType(10), min.GetQuantity())
	assert.Equal(t, items_types.PriceType(500), min.GetPrice())
}
