package depth_test

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	degree = 3
)

func TestLockUnlock(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	d.Lock()
	defer d.Unlock()

	// Add assertions here to verify that the lock and unlock operations are working correctly
	assert.True(t, true)
}

func TestSetAndGetAsk(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	price := items_types.PriceType(100.0)
	quantity := items_types.QuantityType(10.0)

	d.GetAsks().Set(items_types.NewAsk(price, quantity))
	ask := d.GetAsks().Get(items_types.NewAsk(price, quantity))

	// Add assertions here to verify that the ask is set and retrieved correctly
	assert.NotNil(t, ask)
	assert.Equal(t, price, ask.GetDepthItem().GetPrice())
}

func TestSetAndGetBid(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	price := items_types.PriceType(200.0)
	quantity := items_types.QuantityType(20.0)

	d.GetBids().Set(items_types.NewBid(price, quantity))
	bid := d.GetBids().Get(items_types.NewBid(price, quantity))

	// Add assertions here to verify that the bid is set and retrieved correctly
	assert.NotNil(t, bid)
	assert.Equal(t, price, bid.GetDepthItem().GetPrice())
}

// Add more tests here based on your requirements

func initDepths(depth *depth_types.Depths) {
	depth.GetAsks().Set(items_types.NewAsk(1000.0, 10.0))
	depth.GetAsks().Set(items_types.NewAsk(900.0, 20.0))
	depth.GetAsks().Set(items_types.NewAsk(800.0, 30.0))
	depth.GetAsks().Set(items_types.NewAsk(700.0, 20.0))
	depth.GetAsks().Set(items_types.NewAsk(600.0, 10.0))
	depth.GetBids().Set(items_types.NewBid(500.0, 10.0))
	depth.GetBids().Set(items_types.NewBid(400.0, 20.0))
	depth.GetBids().Set(items_types.NewBid(300.0, 30.0))
	depth.GetBids().Set(items_types.NewBid(200.0, 20.0))
	depth.GetBids().Set(items_types.NewBid(100.0, 10.0))
}

func getTestDepths() (asks *btree.BTree, bids *btree.BTree) {
	bids = btree.New(3)
	bidList := []items_types.DepthItem{
		*items_types.New(1.92, 150.2),
		*items_types.New(1.93, 155.4), // local maxima
		*items_types.New(1.94, 150.0),
		*items_types.New(1.941, 130.4),
		*items_types.New(1.947, 172.1),
		*items_types.New(1.948, 187.4),
		*items_types.New(1.949, 236.1), // local maxima
		*items_types.New(1.95, 189.8),
	}
	asks = btree.New(3)
	askList := []items_types.DepthItem{
		*items_types.New(1.951, 217.9), // local maxima
		*items_types.New(1.952, 179.4),
		*items_types.New(1.953, 180.9), // local maxima
		*items_types.New(1.964, 148.5),
		*items_types.New(1.965, 120.0),
		*items_types.New(1.976, 110.0),
		*items_types.New(1.977, 140.0), // local maxima
		*items_types.New(1.988, 90.0),
	}
	for _, bid := range bidList {
		bids.ReplaceOrInsert(&bid)
	}
	for _, ask := range askList {
		asks.ReplaceOrInsert(&ask)
	}

	return
}

func TestGetTargetAsksBidPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(
			d.GetAsks().GetDepths().GetSummaQuantity()*items_types.QuantityType(0.3), depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(
			d.GetBids().GetDepths().GetSummaQuantity()*items_types.QuantityType(0.3), depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaAsks)
		assert.Equal(t, items_types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaBids)
	}()
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(
			d.GetAsks().GetDepths().GetSummaQuantity()*items_types.QuantityType(0.3), depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(
			d.GetBids().GetDepths().GetSummaQuantity()*items_types.QuantityType(0.3), depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaAsks)
		assert.Equal(t, items_types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaBids)
	}()
}

func TestGetMaxAndSummaByQuantityPercent(t *testing.T) {
	// GetMaxAndSummaByQuantityPercent
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantityPercent(30, depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantityPercent(30, depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaAsks)
		assert.Equal(t, items_types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaBids)
	}()
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantityPercent(5, depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantityPercent(5, depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaAsks)
		assert.Equal(t, items_types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), summaBids)
	}()
}

func TestGetAsksBidMaxAndSummaByQuantityPercent(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantityPercent(30, depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantityPercent(30, depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), asks.GetQuantity())
		assert.Equal(t, items_types.QuantityType(10.0), summaAsks)
		assert.Equal(t, items_types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(10.0), bids.GetQuantity())
		assert.Equal(t, items_types.QuantityType(10.0), summaBids)
	}()
	func() {
		asks, summaAsks := d.GetAsks().GetDepths().GetMaxAndSummaByQuantityPercent(40, depths_types.UP)
		bids, summaBids := d.GetBids().GetDepths().GetMaxAndSummaByQuantityPercent(40, depths_types.DOWN)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, items_types.PriceType(700.0), asks.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), asks.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaAsks)
		assert.Equal(t, items_types.PriceType(400.0), bids.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), bids.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaBids)
	}()
}

// func TestGetAsksAndBidsMaxUpToPrice(t *testing.T) {
// 	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
// 	initDepths(d)
// 	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
// 	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetSummaQuantity())
// 	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetSummaQuantity())
// 	maxAsks, maxBids, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByPrice(850.0, 250.0)
// 	assert.Equal(t, items_types.PriceType(800.0), maxAsks.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaAsks)
// 	assert.Equal(t, items_types.PriceType(300.0), maxBids.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaBids)
// 	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(850.0, 250.0, true)
// 	assert.Equal(t, items_types.PriceType(800.0), maxAsks.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaAsks)
// 	assert.Equal(t, items_types.PriceType(300.0), maxBids.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaBids)
// 	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0)
// 	assert.Equal(t, items_types.PriceType(900.0), maxAsks.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(80.0), summaAsks)
// 	assert.Equal(t, items_types.PriceType(200.0), maxBids.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(80.0), summaBids)
// 	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0, true)
// 	assert.Equal(t, items_types.PriceType(800.0), maxAsks.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaAsks)
// 	assert.Equal(t, items_types.PriceType(300.0), maxBids.GetPrice())
// 	assert.Equal(t, items_types.QuantityType(60.0), summaBids)
// }

func TestGetFilteredByPercentAsks(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentAsks method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	filtered, summa, max, min := d.GetAsks().GetDepths().GetFilteredByPercent(func(i *items_types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetAsks().GetDepths().GetSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, items_types.QuantityType(30.0), summa)
	assert.Equal(t, items_types.QuantityType(30.0), max)
	assert.Equal(t, items_types.QuantityType(30.0), min)

	filtered, summa, max, min = d.GetAsks().GetDepths().GetFilteredByPercent(func(i *items_types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetAsks().GetDepths().GetSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, items_types.QuantityType(60.0), summa)
	assert.Equal(t, items_types.QuantityType(20.0), max)
	assert.Equal(t, items_types.QuantityType(10.0), min)
}

func TestGetFilteredByPercentBids(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentBids method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	filtered, summa, max, min := d.GetBids().GetDepths().GetFilteredByPercent(func(i *items_types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetBids().GetDepths().GetSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, items_types.QuantityType(30.0), summa)
	assert.Equal(t, items_types.QuantityType(30.0), max)
	assert.Equal(t, items_types.QuantityType(30.0), min)

	filtered, summa, max, min = d.GetBids().GetDepths().GetFilteredByPercent(func(i *items_types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetBids().GetDepths().GetSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, items_types.QuantityType(60.0), summa)
	assert.Equal(t, items_types.QuantityType(20.0), max)
	assert.Equal(t, items_types.QuantityType(10.0), min)
}

func TestGetSummaOfAsksAndBidFromRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetSummaOfAsksFromRange method works correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	summaAsk, max, min := d.GetAsks().GetDepths().GetSummaByRange(600.0, 800.0, func(d *items_types.DepthItem) bool { return true })
	assert.Equal(t, items_types.QuantityType(50.0), summaAsk)
	assert.Equal(t, items_types.QuantityType(30.0), max)
	assert.Equal(t, items_types.QuantityType(20.0), min)
	summaBid, max, min := d.GetBids().GetDepths().GetSummaByRange(300.0, 50.0, func(d *items_types.DepthItem) bool { return true })
	assert.Equal(t, items_types.QuantityType(30.0), summaBid)
	assert.Equal(t, items_types.QuantityType(20.0), max)
	assert.Equal(t, items_types.QuantityType(10.0), min)
}

// func TestMinMax(t *testing.T) {
// 	d := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
// 	initDepths(d)
// 	// Add assertions here to verify that the Min and Max methods work correctly
// 	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
// 	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
// 	min, err := d.AskMin()
// 	assert.Nil(t, err)
// 	assert.Equal(t, items_types.PriceType(600.0), min.GetPrice())
// 	max, err := d.AskMax()
// 	assert.Nil(t, err)
// 	assert.Equal(t, items_types.PriceType(800.0), max.GetPrice())
// 	min, err = d.BidMin()
// 	assert.Nil(t, err)
// 	assert.Equal(t, items_types.PriceType(500.0), min.GetPrice())
// 	max, err = d.BidMax()
// 	assert.Nil(t, err)
// 	assert.Equal(t, items_types.PriceType(300.0), max.GetPrice())
// }

func TestGetAsksAndBidSummaAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	func() {
		maxAsk1, summaAsks1 := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(40, depths_types.UP)
		maxBid1, summaBids1 := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(40, depths_types.DOWN)
		assert.Equal(t, items_types.PriceType(700.0), maxAsk1.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxAsk1.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaAsks1)
		assert.Equal(t, items_types.PriceType(400.0), maxBid1.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxBid1.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaBids1)
		maxAsk3, summaAsks3 := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(85.0, depths_types.UP)
		maxBid3, summaBids3 := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(85.0, depths_types.DOWN)
		assert.Equal(t, items_types.PriceType(900.0), maxAsk3.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxAsk3.GetQuantity())
		assert.Equal(t, items_types.QuantityType(80.0), summaAsks3)
		assert.Equal(t, items_types.PriceType(200.0), maxBid3.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxBid3.GetQuantity())
		assert.Equal(t, items_types.QuantityType(80.0), summaBids3)
		summaAsks2, maxAsks2, minAsks2 := d.GetAsks().GetDepths().GetSummaByRange(maxAsk1.GetPrice(), maxAsk3.GetPrice())
		summaBids2, maxBid2, minBid2 := d.GetBids().GetDepths().GetSummaByRange(maxBid1.GetPrice(), maxBid3.GetPrice())
		assert.Equal(t, items_types.QuantityType(30.0), maxAsks2)
		assert.Equal(t, items_types.QuantityType(20.0), minAsks2)
		assert.Equal(t, items_types.QuantityType(50.0), summaAsks2)
		assert.Equal(t, items_types.QuantityType(30.0), maxBid2)
		assert.Equal(t, items_types.QuantityType(20.0), minBid2)
		assert.Equal(t, items_types.QuantityType(50.0), summaBids2)
		assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
		assert.Equal(t, summaBids2, summaBids3-summaBids1)
	}()
	func() {
		maxAsk1, summaAsks1 := d.GetAsks().GetDepths().GetMaxAndSummaByPrice(700.0, depths_types.UP)
		maxBid1, summaBids1 := d.GetBids().GetDepths().GetMaxAndSummaByPrice(400.0, depths_types.DOWN)
		assert.Equal(t, items_types.PriceType(700.0), maxAsk1.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxAsk1.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaAsks1)
		assert.Equal(t, items_types.PriceType(400.0), maxBid1.GetPrice())
		assert.Equal(t, items_types.QuantityType(20.0), maxBid1.GetQuantity())
		assert.Equal(t, items_types.QuantityType(30.0), summaBids1)
		maxAsk3, summaAsks3 := d.GetAsks().GetDepths().GetMaxAndSummaByPrice(850.0, depths_types.UP)
		maxBid3, summaBids3 := d.GetBids().GetDepths().GetMaxAndSummaByPrice(250.0, depths_types.DOWN)
		assert.Equal(t, items_types.PriceType(800.0), maxAsk3.GetPrice())
		assert.Equal(t, items_types.QuantityType(30.0), maxAsk3.GetQuantity())
		assert.Equal(t, items_types.QuantityType(60.0), summaAsks3)
		assert.Equal(t, items_types.PriceType(300.0), maxBid3.GetPrice())
		assert.Equal(t, items_types.QuantityType(30.0), maxBid3.GetQuantity())
		assert.Equal(t, items_types.QuantityType(60.0), summaBids3)
		summaAsks2, maxAsks2, minAsks2 := d.GetAsks().GetDepths().GetSummaByRange(maxAsk1.GetPrice(), maxAsk3.GetPrice())
		summaBids2, maxBid2, minBid2 := d.GetBids().GetDepths().GetSummaByRange(maxBid1.GetPrice(), maxBid3.GetPrice())
		assert.Equal(t, items_types.QuantityType(30.0), maxAsks2)
		assert.Equal(t, items_types.QuantityType(30.0), minAsks2)
		assert.Equal(t, items_types.QuantityType(30.0), summaAsks2)
		assert.Equal(t, items_types.QuantityType(30.0), maxBid2)
		assert.Equal(t, items_types.QuantityType(30.0), minBid2)
		assert.Equal(t, items_types.QuantityType(30.0), summaBids2)
		assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
		assert.Equal(t, summaBids2, summaBids3-summaBids1)
	}()
}

func TestGetTargetAsksBidPriceAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, items_types.QuantityType(90.0), d.GetAsks().GetDepths().GetSummaQuantity())
	assert.Equal(t, items_types.QuantityType(90.0), d.GetBids().GetDepths().GetSummaQuantity())
	ask1, summaAsks1 := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(20, depths_types.UP)
	bid1, summaBids1 := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(20, depths_types.DOWN)
	ask2, summaAsks3 := d.GetAsks().GetDepths().GetMaxAndSummaByQuantity(50, depths_types.UP)
	bid2, summaBids3 := d.GetBids().GetDepths().GetMaxAndSummaByQuantity(50, depths_types.DOWN)
	assert.Equal(t, items_types.QuantityType(10.0), summaAsks1)
	assert.Equal(t, items_types.QuantityType(10.0), summaBids1)
	assert.Equal(t, items_types.QuantityType(30.0), summaAsks3)
	assert.Equal(t, items_types.QuantityType(30.0), summaBids3)
	summaAsks2, max, min := d.GetAsks().GetDepths().GetSummaByRange(ask1.GetPrice(), ask2.GetPrice())
	assert.Equal(t, items_types.QuantityType(20.0), max)
	assert.Equal(t, items_types.QuantityType(20.0), min)
	summaBids2, max, min := d.GetBids().GetDepths().GetSummaByRange(bid1.GetPrice(), bid2.GetPrice())
	assert.Equal(t, items_types.QuantityType(20.0), max)
	assert.Equal(t, items_types.QuantityType(20.0), min)
	assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
	assert.Equal(t, summaBids2, summaBids3-summaBids1)
}

func TestGetTargetPrices(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetPrices method works correctly
	ask1, bid1, summaAsks1, summaBids1 := d.GetTargetPrices(20)
	ask2, bid2, summaAsks2, summaBids2 := d.GetTargetPrices(50)
	assert.Equal(t, items_types.PriceType(600.0), ask1)
	assert.Equal(t, items_types.PriceType(500.0), bid1)
	assert.Equal(t, items_types.QuantityType(10.0), summaAsks1)
	assert.Equal(t, items_types.QuantityType(10.0), summaBids1)
	assert.Equal(t, items_types.PriceType(700.0), ask2)
	assert.Equal(t, items_types.PriceType(400.0), bid2)
	assert.Equal(t, items_types.QuantityType(30.0), summaAsks2)
	assert.Equal(t, items_types.QuantityType(30.0), summaBids2)
}

func TestGetLimitPrices(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetPrices method works correctly
	ask1, bid1, summaAsks1, summaBids1 := d.GetLimitPrices()
	assert.Equal(t, items_types.PriceType(800.0), ask1)
	assert.Equal(t, items_types.PriceType(300.0), bid1)
	assert.Equal(t, items_types.QuantityType(60.0), summaAsks1)
	assert.Equal(t, items_types.QuantityType(60.0), summaBids1)
}

func TestNew(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depths_types.DepthStreamRate100ms)
	// Add assertions here to verify that the New method works correctly
	assert.NotNil(t, d)
	assert.Equal(t, "BTCUSDT", d.Symbol())
	assert.Equal(t, depths_types.DepthAPILimit(100), d.GetLimitDepth())
	assert.Equal(t, depths_types.DepthStreamLevel(20), d.GetLimitStream())
	assert.Equal(t, depths_types.DepthStreamRate100ms, d.GetRateStream())
}

func TestGetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", false, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetAsks().SetTree(asks)
	ask := ds.GetAsks().Get(items_types.NewAsk(1.951))
	if ask == nil {
		t.Errorf("Failed to get ask")
	}
	ask = ds.GetAsks().Get(items_types.NewAsk(0))
	if ask != nil {
		t.Errorf("Failed to get ask")
	}
}

func TestGetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetBids().SetTree(bids)
	bid := ds.GetBids().Get(items_types.NewBid(1.93))
	if bid == nil {
		t.Errorf("Failed to get bid")
	}
}

func TestSetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetAsks().SetTree(asks)
	ask := items_types.NewAsk(1.96, 200.0)
	ds.GetAsks().Set(ask)
	if ds.GetAsks().Get(items_types.NewAsk(1.96)) == nil {
		t.Errorf("Failed to set ask")
	}
}

func TestSetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetBids().SetTree(bids)
	bid := items_types.NewBid(1.96, 200.0)
	ds.GetBids().Set(bid)
	if ds.GetBids().Get(items_types.NewBid(1.96)) == nil {
		t.Errorf("Failed to set bid")
	}
}

func TestRestrictAskAndBidDown(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetAsks().SetTree(asks)
	ds.GetBids().SetTree(bids)
	ds.GetAsks().GetDepths().RestrictDown(1.957)
	ds.GetBids().GetDepths().RestrictDown(1.949)
	if ds.GetAsks().Get(items_types.NewAsk(1.951)) != nil {
		t.Errorf("Failed to restrict ask")
	}
	if ds.GetBids().Get(items_types.NewBid(1.93)) != nil {
		t.Errorf("Failed to restrict bid")
	}
}

func TestRestrictAskAndBidUp(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetAsks().SetTree(asks)
	ds.GetBids().SetTree(bids)
	ds.GetAsks().GetDepths().RestrictDown(1.957)
	ds.GetBids().GetDepths().RestrictUp(1.949)
	if ds.GetAsks().Get(items_types.NewAsk(1.951)) != nil {
		t.Errorf("Failed to restrict ask")
	}
	if ds.GetBids().Get(items_types.NewBid(1.95)) != nil {
		t.Errorf("Failed to restrict bid")
	}
}

func summaAsksAndBids(ds *depth_types.Depths) (summaAsks, summaBids items_types.QuantityType) {
	ds.GetAsks().GetTree().Ascend(func(i btree.Item) bool {
		summaAsks += i.(*items_types.DepthItem).GetQuantity()
		return true
	})
	ds.GetBids().GetTree().Ascend(func(i btree.Item) bool {
		summaBids += i.(*items_types.DepthItem).GetQuantity()
		return true
	})
	return
}

func TestUpdateAskAndBid(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetAsks().SetTree(asks)
	ds.GetBids().SetTree(bids)
	ask := ds.GetAsks().Get(items_types.NewAsk(1.951))
	bid := ds.GetBids().Get(items_types.NewBid(1.951))
	summaAsks, summaBids := summaAsksAndBids(ds)
	assert.Equal(t, items_types.QuantityType(217.9), ask.GetDepthItem().GetQuantity())
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetDepths().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetDepths().GetSummaQuantity()), 6))
	ds.UpdateAsk(items_types.NewAsk(1.951, 300.0))
	ask = ds.GetAsks().Get(items_types.NewAsk(1.951))
	bid = ds.GetBids().Get(items_types.NewBid(1.951))
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, items_types.QuantityType(300.0), ask.GetDepthItem().GetQuantity())
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetDepths().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetDepths().GetSummaQuantity()), 6))

	ds.UpdateBid(items_types.NewBid(1.951, 300.0))
	ask = ds.GetAsks().Get(items_types.NewAsk(1.951))
	bid = ds.GetBids().Get(items_types.NewBid(1.951))
	assert.Nil(t, ask)
	assert.Equal(t, items_types.QuantityType(300.0), bid.GetDepthItem().GetQuantity())
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetDepths().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetDepths().GetSummaQuantity()), 6))
	ds.GetBids().Set(items_types.NewBid(2.0, 100))
	assert.Equal(t, items_types.QuantityType(1771.3999999999999), ds.GetBids().GetDepths().GetSummaQuantity())
	ds.GetBids().Delete(items_types.NewBid(2.0))
	assert.Equal(t, items_types.QuantityType(1671.3999999999999), ds.GetBids().GetDepths().GetSummaQuantity())
	ds.GetBids().Delete(items_types.NewBid(2.0))
	assert.Equal(t, items_types.QuantityType(1671.3999999999999), ds.GetBids().GetDepths().GetSummaQuantity())
}

func TestGetFilteredByPercentAsksAndBids(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
	ds.GetBids().SetTree(bids)
	ds.GetAsks().SetTree(asks)
	normalizedAsks, _, _, _ := ds.GetAsks().GetDepths().GetFilteredByPercent()
	normalizedBids, _, _, _ := ds.GetBids().GetDepths().GetFilteredByPercent()
	assert.NotNil(t, normalizedAsks)
	assert.NotNil(t, normalizedBids)
	normalizedAsksArray := make([]items_types.DepthItem, 0)
	normalizedBidsArray := make([]items_types.DepthItem, 0)
	normalizedAsks.Ascend(func(i btree.Item) bool {
		normalizedAsksArray = append(normalizedAsksArray, *i.(*items_types.DepthItem))
		return true
	})
	normalizedBids.Ascend(func(i btree.Item) bool {
		normalizedBidsArray = append(normalizedBidsArray, *i.(*items_types.DepthItem))
		return true
	})
	assert.Equal(t, 8, len(normalizedAsksArray))
	assert.Equal(t, 8, len(normalizedBidsArray))
}

// func TestDepthInterface(t *testing.T) {
// 	test := func(ds depth_interface.Depth) {
// 		ds.UpdateBid(1.93, 300.0)
// 		bid := ds.GetBid(1.93)
// 		assert.Equal(t, items_types.QuantityType(300.0), bid.(*items_types.DepthItem).GetQuantity())
// 		ds.UpdateAsk(1.951, 300.0)
// 		ask := ds.GetAsk(1.951)
// 		assert.Equal(t, items_types.QuantityType(300.0), ask.(*items_types.DepthItem).GetQuantity())
// 	}
// 	asks, bids := getTestDepths()
// 	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
// 	ds.GetBids().SetTree(bids)
// 	ds.GetAsks().SetTree(asks)
// 	test(ds)
// }

func TestAsksAndBidMiddleQuantity(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		initDepths(ds)
		asksMiddle := ds.GetAsks().GetDepths().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(18.0), asksMiddle)
		bidsMiddle := ds.GetBids().GetDepths().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(18.0), bidsMiddle)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().SetTree(asks)
		ds.GetBids().SetTree(bids)
		asksMiddle := ds.GetAsks().GetDepths().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(148.3375), asksMiddle)
		bidsMiddle := ds.GetBids().GetDepths().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(171.42499999999998), bidsMiddle)
	}()
}

func TestAsksAndBidStandardDeviation(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		initDepths(ds)
		asksSquares := ds.GetAsks().GetDepths().GetStandardDeviation()
		assert.Equal(t, 7.483314773547883, asksSquares)
		bidsSquares := ds.GetBids().GetDepths().GetStandardDeviation()
		assert.Equal(t, 7.483314773547883, bidsSquares)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().SetTree(asks)
		ds.GetBids().SetTree(bids)
		asksSquares := ds.GetAsks().GetDepths().GetStandardDeviation()
		assert.Equal(t, 39.70157230828522, asksSquares)
		bidsSquares := ds.GetBids().GetDepths().GetStandardDeviation()
		assert.Equal(t, 30.873805644915237, bidsSquares)
	}()
}

func TestAskAndBidDelete(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().Set(items_types.NewAsk(800, 100))
		ds.GetAsks().Set(items_types.NewAsk(750, 150))
		ds.GetAsks().Delete(items_types.NewAsk(800))
		ask := ds.GetAsks().Get(items_types.NewAsk(800))
		assert.Nil(t, ask)
		assert.Equal(t, 1, ds.GetAsks().GetTree().Len())
		ds.GetAsks().Delete(items_types.NewAsk(750))
		ask = ds.GetAsks().Get(items_types.NewAsk(750))
		assert.Nil(t, ask)
		assert.Equal(t, 0, ds.GetAsks().GetTree().Len())
	}()
}

func TestAskAndBidSummaQuantity(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().Set(items_types.NewAsk(800, 100))
		assert.Equal(t, items_types.QuantityType(100.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(790, 100))
		assert.Equal(t, items_types.QuantityType(200.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(780, 100))
		assert.Equal(t, items_types.QuantityType(300.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(770, 100))
		assert.Equal(t, items_types.QuantityType(400.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(760, 100))
		assert.Equal(t, items_types.QuantityType(500.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(750, 100))
		assert.Equal(t, items_types.QuantityType(600.0), ds.GetAsks().GetDepths().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(740, 100))
		assert.Equal(t, items_types.QuantityType(700.0), ds.GetAsks().GetDepths().GetSummaQuantity())
	}()
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depths_types.DepthStreamRate100ms)
		ds.GetBids().Set(items_types.NewBid(800, 100))
		assert.Equal(t, items_types.QuantityType(100.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(790, 100))
		assert.Equal(t, items_types.QuantityType(200.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(780, 100))
		assert.Equal(t, items_types.QuantityType(300.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(770, 100))
		assert.Equal(t, items_types.QuantityType(400.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(760, 100))
		assert.Equal(t, items_types.QuantityType(500.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(750, 100))
		assert.Equal(t, items_types.QuantityType(600.0), ds.GetBids().GetDepths().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(740, 100))
		assert.Equal(t, items_types.QuantityType(700.0), ds.GetBids().GetDepths().GetSummaQuantity())
	}()
}

// func TestAskAndBidMinMaxQuantity(t *testing.T) {
// 	func() {
// 		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depths_types.DepthStreamRate100ms)
// 		minTest := func() items_types.QuantityType {
// 			min, err := ds.AskMin()
// 			if err != nil {
// 				return 0
// 			}
// 			return min.GetQuantity()
// 		}
// 		maxTest := func() items_types.QuantityType {
// 			max, err := ds.AskMax()
// 			if err != nil {
// 				return 0
// 			}
// 			return max.GetQuantity()
// 		}
// 		ds.SetAsk(800, 100)
// 		assert.Equal(t, items_types.QuantityType(100.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(100.0), maxTest())
// 		ds.SetAsk(790, 200)
// 		assert.Equal(t, items_types.QuantityType(100.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(200.0), maxTest())
// 		ds.SetAsk(780, 300)
// 		assert.Equal(t, items_types.QuantityType(100.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(800)
// 		assert.Equal(t, items_types.QuantityType(200.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(790)
// 		assert.Equal(t, items_types.QuantityType(300.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(780)
// 		assert.Equal(t, items_types.QuantityType(0.0), minTest())
// 		assert.Equal(t, items_types.QuantityType(0.0), maxTest())
// 	}()
// }
