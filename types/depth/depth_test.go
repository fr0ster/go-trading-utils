package depth_test

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	types "github.com/fr0ster/go-trading-utils/types/depth/types"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	degree = 3
)

func TestLockUnlock(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	d.Lock()
	defer d.Unlock()

	// Add assertions here to verify that the lock and unlock operations are working correctly
	assert.True(t, true)
}

func TestSetAndGetAsk(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	price := types.PriceType(100.0)
	quantity := types.QuantityType(10.0)

	d.SetAsk(price, quantity)
	ask := d.GetAsk(price)

	// Add assertions here to verify that the ask is set and retrieved correctly
	assert.NotNil(t, ask)
	assert.Equal(t, price, ask.(*types.DepthItem).GetPrice())
}

func TestSetAndGetBid(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	price := types.PriceType(200.0)
	quantity := types.QuantityType(20.0)

	d.SetBid(price, quantity)
	bid := d.GetBid(price)

	// Add assertions here to verify that the bid is set and retrieved correctly
	assert.NotNil(t, bid)
	assert.Equal(t, price, bid.(*types.DepthItem).GetPrice())
}

// Add more tests here based on your requirements

func initDepths(depth *depth_types.Depth) {
	depth.SetAsk(1000.0, 10.0)
	depth.SetAsk(900.0, 20.0)
	depth.SetAsk(800.0, 30.0)
	depth.SetAsk(700.0, 20.0)
	depth.SetAsk(600.0, 10.0)
	depth.SetBid(500.0, 10.0)
	depth.SetBid(400.0, 20.0)
	depth.SetBid(300.0, 30.0)
	depth.SetBid(200.0, 20.0)
	depth.SetBid(100.0, 10.0)
}

func getTestDepths() (asks *btree.BTree, bids *btree.BTree) {
	bids = btree.New(3)
	bidList := []types.DepthItem{
		*types.NewDepthItem(1.92, 150.2),
		*types.NewDepthItem(1.93, 155.4), // local maxima
		*types.NewDepthItem(1.94, 150.0),
		*types.NewDepthItem(1.941, 130.4),
		*types.NewDepthItem(1.947, 172.1),
		*types.NewDepthItem(1.948, 187.4),
		*types.NewDepthItem(1.949, 236.1), // local maxima
		*types.NewDepthItem(1.95, 189.8),
	}
	asks = btree.New(3)
	askList := []types.DepthItem{
		*types.NewDepthItem(1.951, 217.9), // local maxima
		*types.NewDepthItem(1.952, 179.4),
		*types.NewDepthItem(1.953, 180.9), // local maxima
		*types.NewDepthItem(1.964, 148.5),
		*types.NewDepthItem(1.965, 120.0),
		*types.NewDepthItem(1.976, 110.0),
		*types.NewDepthItem(1.977, 140.0), // local maxima
		*types.NewDepthItem(1.988, 90.0),
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
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	func() {
		asks, bids, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByQuantity(d.GetAsksSummaQuantity()*types.QuantityType(0.3), d.GetBidsSummaQuantity()*0.3)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, types.QuantityType(10.0), summaAsks)
		assert.Equal(t, types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, types.QuantityType(10.0), summaBids)
	}()
	func() {
		asks, bids, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByQuantity(d.GetAsksSummaQuantity()*0.3, d.GetBidsSummaQuantity()*0.3, true)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, types.PriceType(600.0), asks.GetPrice())
		assert.Equal(t, types.QuantityType(10.0), summaAsks)
		assert.Equal(t, types.PriceType(500.0), bids.GetPrice())
		assert.Equal(t, types.QuantityType(10.0), summaBids)
	}()
}

// func TestGetAsksBidMaxAndSummaByQuantityPercent(t *testing.T) {
// 	d := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
// 	initDepths(d)
// 	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
// 	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
// 	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
// 	func() {
// 		asks, bids, summaAsks, summaBids, err := d.GetAsksBidMaxAndSummaByQuantityPercent(10, 10)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, asks)
// 		assert.NotNil(t, bids)
// 		assert.Equal(t, types.PriceType(600.0), asks.GetPrice())
// 		assert.Equal(t, types.QuantityType(10.0), asks.GetQuantity())
// 		assert.Equal(t, types.QuantityType(10.0), summaAsks)
// 		assert.Equal(t, types.PriceType(500.0), bids.GetPrice())
// 		assert.Equal(t, types.QuantityType(10.0), bids.GetQuantity())
// 		assert.Equal(t, types.QuantityType(10.0), summaBids)
// 	}()
// 	func() {
// 		asks, bids, summaAsks, summaBids, err := d.GetAsksBidMaxAndSummaByQuantityPercent(40, 40)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, asks)
// 		assert.NotNil(t, bids)
// 		assert.Equal(t, types.PriceType(700.0), asks.GetPrice())
// 		assert.Equal(t, types.QuantityType(20.0), asks.GetQuantity())
// 		assert.Equal(t, types.QuantityType(30.0), summaAsks)
// 		assert.Equal(t, types.PriceType(400.0), bids.GetPrice())
// 		assert.Equal(t, types.QuantityType(20.0), bids.GetQuantity())
// 		assert.Equal(t, types.QuantityType(30.0), summaBids)
// 	}()
// }

func TestGetAsksAndBidsMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	maxAsks, maxBids, summaAsks, summaBids := d.GetAsksBidMaxAndSummaByPrice(850.0, 250.0)
	assert.Equal(t, types.PriceType(800.0), maxAsks.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaAsks)
	assert.Equal(t, types.PriceType(300.0), maxBids.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaBids)
	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(850.0, 250.0, true)
	assert.Equal(t, types.PriceType(800.0), maxAsks.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaAsks)
	assert.Equal(t, types.PriceType(300.0), maxBids.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaBids)
	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0)
	assert.Equal(t, types.PriceType(900.0), maxAsks.GetPrice())
	assert.Equal(t, types.QuantityType(80.0), summaAsks)
	assert.Equal(t, types.PriceType(200.0), maxBids.GetPrice())
	assert.Equal(t, types.QuantityType(80.0), summaBids)
	maxAsks, maxBids, summaAsks, summaBids = d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0, true)
	assert.Equal(t, types.PriceType(800.0), maxAsks.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaAsks)
	assert.Equal(t, types.PriceType(300.0), maxBids.GetPrice())
	assert.Equal(t, types.QuantityType(60.0), summaBids)
}

func TestGetFilteredByPercentAsks(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentAsks method works correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	filtered, summa, max, min := d.GetFilteredByPercentAsks(func(i *types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetAsksSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, types.QuantityType(30.0), summa)
	assert.Equal(t, types.QuantityType(30.0), max)
	assert.Equal(t, types.QuantityType(30.0), min)

	filtered, summa, max, min = d.GetFilteredByPercentAsks(func(i *types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetAsksSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, types.QuantityType(60.0), summa)
	assert.Equal(t, types.QuantityType(20.0), max)
	assert.Equal(t, types.QuantityType(10.0), min)
}

func TestGetFilteredByPercentBids(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentBids method works correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	filtered, summa, max, min := d.GetFilteredByPercentBids(func(i *types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetBidsSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, types.QuantityType(30.0), summa)
	assert.Equal(t, types.QuantityType(30.0), max)
	assert.Equal(t, types.QuantityType(30.0), min)

	filtered, summa, max, min = d.GetFilteredByPercentBids(func(i *types.DepthItem) bool {
		return i.GetQuantity()*100/d.GetBidsSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, types.QuantityType(60.0), summa)
	assert.Equal(t, types.QuantityType(20.0), max)
	assert.Equal(t, types.QuantityType(10.0), min)
}

func TestGetSummaOfAsksAndBidFromRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetSummaOfAsksFromRange method works correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	summaAsk, max, min := d.GetSummaOfAsksFromRange(600.0, 800.0, func(d *types.DepthItem) bool { return true })
	assert.Equal(t, types.QuantityType(50.0), summaAsk)
	assert.Equal(t, types.QuantityType(30.0), max)
	assert.Equal(t, types.QuantityType(20.0), min)
	summaBid, max, min := d.GetSummaOfBidsFromRange(300.0, 50.0, func(d *types.DepthItem) bool { return true })
	assert.Equal(t, types.QuantityType(30.0), summaBid)
	assert.Equal(t, types.QuantityType(20.0), max)
	assert.Equal(t, types.QuantityType(10.0), min)
}

// func TestMinMax(t *testing.T) {
// 	d := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
// 	initDepths(d)
// 	// Add assertions here to verify that the Min and Max methods work correctly
// 	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
// 	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
// 	min, err := d.AskMin()
// 	assert.Nil(t, err)
// 	assert.Equal(t, types.PriceType(600.0), min.GetPrice())
// 	max, err := d.AskMax()
// 	assert.Nil(t, err)
// 	assert.Equal(t, types.PriceType(800.0), max.GetPrice())
// 	min, err = d.BidMin()
// 	assert.Nil(t, err)
// 	assert.Equal(t, types.PriceType(500.0), min.GetPrice())
// 	max, err = d.BidMax()
// 	assert.Nil(t, err)
// 	assert.Equal(t, types.PriceType(300.0), max.GetPrice())
// }

func TestGetAsksAndBidSummaAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	func() {
		maxAsk1, maxBid1, summaAsks1, summaBids1 := d.GetAsksBidMaxAndSummaByPrice(700.0, 400.0)
		assert.Equal(t, types.PriceType(700.0), maxAsk1.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxAsk1.GetQuantity())
		assert.Equal(t, types.QuantityType(30.0), summaAsks1)
		assert.Equal(t, types.PriceType(400.0), maxBid1.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxBid1.GetQuantity())
		assert.Equal(t, types.QuantityType(30.0), summaBids1)
		maxAsk3, maxBid3, summaAsks3, summaBids3 := d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0)
		assert.Equal(t, types.PriceType(900.0), maxAsk3.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxAsk3.GetQuantity())
		assert.Equal(t, types.QuantityType(80.0), summaAsks3)
		assert.Equal(t, types.PriceType(200.0), maxBid3.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxBid3.GetQuantity())
		assert.Equal(t, types.QuantityType(80.0), summaBids3)
		summaAsks2, maxAsks2, minAsks2 := d.GetSummaOfAsksFromRange(maxAsk1.GetPrice(), maxAsk3.GetPrice())
		summaBids2, maxBid2, minBid2 := d.GetSummaOfBidsFromRange(maxBid1.GetPrice(), maxBid3.GetPrice())
		assert.Equal(t, types.QuantityType(30.0), maxAsks2)
		assert.Equal(t, types.QuantityType(20.0), minAsks2)
		assert.Equal(t, types.QuantityType(50.0), summaAsks2)
		assert.Equal(t, types.QuantityType(30.0), maxBid2)
		assert.Equal(t, types.QuantityType(20.0), minBid2)
		assert.Equal(t, types.QuantityType(50.0), summaBids2)
		assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
		assert.Equal(t, summaBids2, summaBids3-summaBids1)
	}()
	func() {
		maxAsk1, maxBid1, summaAsks1, summaBids1 := d.GetAsksBidMaxAndSummaByPrice(700.0, 400.0, true)
		assert.Equal(t, types.PriceType(700.0), maxAsk1.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxAsk1.GetQuantity())
		assert.Equal(t, types.QuantityType(30.0), summaAsks1)
		assert.Equal(t, types.PriceType(400.0), maxBid1.GetPrice())
		assert.Equal(t, types.QuantityType(20.0), maxBid1.GetQuantity())
		assert.Equal(t, types.QuantityType(30.0), summaBids1)
		maxAsk3, maxBid3, summaAsks3, summaBids3 := d.GetAsksBidMaxAndSummaByPrice(950.0, 150.0, true)
		assert.Equal(t, types.PriceType(800.0), maxAsk3.GetPrice())
		assert.Equal(t, types.QuantityType(30.0), maxAsk3.GetQuantity())
		assert.Equal(t, types.QuantityType(60.0), summaAsks3)
		assert.Equal(t, types.PriceType(300.0), maxBid3.GetPrice())
		assert.Equal(t, types.QuantityType(30.0), maxBid3.GetQuantity())
		assert.Equal(t, types.QuantityType(60.0), summaBids3)
		summaAsks2, maxAsks2, minAsks2 := d.GetSummaOfAsksFromRange(maxAsk1.GetPrice(), maxAsk3.GetPrice())
		summaBids2, maxBid2, minBid2 := d.GetSummaOfBidsFromRange(maxBid1.GetPrice(), maxBid3.GetPrice())
		assert.Equal(t, types.QuantityType(30.0), maxAsks2)
		assert.Equal(t, types.QuantityType(30.0), minAsks2)
		assert.Equal(t, types.QuantityType(30.0), summaAsks2)
		assert.Equal(t, types.QuantityType(30.0), maxBid2)
		assert.Equal(t, types.QuantityType(30.0), minBid2)
		assert.Equal(t, types.QuantityType(30.0), summaBids2)
		assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
		assert.Equal(t, summaBids2, summaBids3-summaBids1)
	}()
}

func TestGetTargetAsksBidPriceAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, types.QuantityType(90.0), d.GetAsksSummaQuantity())
	assert.Equal(t, types.QuantityType(90.0), d.GetBidsSummaQuantity())
	ask1, bid1, summaAsks1, summaBids1 := d.GetAsksBidMaxAndSummaByQuantity(20, 20)
	ask2, bid2, summaAsks3, summaBids3 := d.GetAsksBidMaxAndSummaByQuantity(50, 50)
	assert.Equal(t, types.QuantityType(10.0), summaAsks1)
	assert.Equal(t, types.QuantityType(10.0), summaBids1)
	assert.Equal(t, types.QuantityType(30.0), summaAsks3)
	assert.Equal(t, types.QuantityType(30.0), summaBids3)
	summaAsks2, max, min := d.GetSummaOfAsksFromRange(ask1.GetPrice(), ask2.GetPrice())
	assert.Equal(t, types.QuantityType(20.0), max)
	assert.Equal(t, types.QuantityType(20.0), min)
	summaBids2, max, min := d.GetSummaOfBidsFromRange(bid1.GetPrice(), bid2.GetPrice())
	assert.Equal(t, types.QuantityType(20.0), max)
	assert.Equal(t, types.QuantityType(20.0), min)
	assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
	assert.Equal(t, summaBids2, summaBids3-summaBids1)
}

func TestGetTargetPrices(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetPrices method works correctly
	ask1, bid1, summaAsks1, summaBids1 := d.GetTargetPrices(20)
	ask2, bid2, summaAsks2, summaBids2 := d.GetTargetPrices(50)
	assert.Equal(t, types.PriceType(600.0), ask1)
	assert.Equal(t, types.PriceType(500.0), bid1)
	assert.Equal(t, types.QuantityType(10.0), summaAsks1)
	assert.Equal(t, types.QuantityType(10.0), summaBids1)
	assert.Equal(t, types.PriceType(700.0), ask2)
	assert.Equal(t, types.PriceType(400.0), bid2)
	assert.Equal(t, types.QuantityType(30.0), summaAsks2)
	assert.Equal(t, types.QuantityType(30.0), summaBids2)
}

func TestNew(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 100, 2, depth_types.DepthStreamRate100ms)
	// Add assertions here to verify that the New method works correctly
	assert.NotNil(t, d)
	assert.Equal(t, "BTCUSDT", d.Symbol())
	assert.Equal(t, depth_types.DepthAPILimit(100), d.GetLimitDepth())
	assert.Equal(t, depth_types.DepthStreamLevel(20), d.GetLimitStream())
	assert.Equal(t, depth_types.DepthStreamRate100ms, d.GetRateStream())
}

func TestGetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ask := ds.GetAsk(1.951)
	if ask == nil {
		t.Errorf("Failed to get ask")
	}
	ask = ds.GetAsk(0)
	if ask != nil {
		t.Errorf("Failed to get ask")
	}
}

func TestGetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	bid := ds.GetBid(1.93)
	if bid == nil {
		t.Errorf("Failed to get bid")
	}
}

func TestSetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ask := types.NewDepthItem(1.96, 200.0)
	ds.SetAsk(ask.GetPrice(), ask.GetQuantity())
	if ds.GetAsk(1.96) == nil {
		t.Errorf("Failed to set ask")
	}
}

func TestSetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	bid := types.NewDepthItem(1.96, 200.0)
	ds.SetBid(bid.GetPrice(), bid.GetQuantity())
	if ds.GetBid(1.96) == nil {
		t.Errorf("Failed to set bid")
	}
}

func TestRestrictAskAndBidDown(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ds.SetBids(bids)
	ds.RestrictAskDown(1.957)
	ds.RestrictBidDown(1.949)
	if ds.GetAsk(1.951) != nil {
		t.Errorf("Failed to restrict ask")
	}
	if ds.GetBid(1.93) != nil {
		t.Errorf("Failed to restrict bid")
	}
}

func TestRestrictAskAndBidUp(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ds.SetBids(bids)
	ds.RestrictAskUp(1.952)
	ds.RestrictBidUp(1.93)
	if ds.GetAsk(1.953) != nil {
		t.Errorf("Failed to restrict ask")
	}
	if ds.GetBid(1.931) != nil {
		t.Errorf("Failed to restrict bid")
	}
}

func summaAsksAndBids(ds *depth_types.Depth) (summaAsks, summaBids types.QuantityType) {
	ds.GetAsks().Ascend(func(i btree.Item) bool {
		summaAsks += i.(*types.DepthItem).GetQuantity()
		return true
	})
	ds.GetBids().Ascend(func(i btree.Item) bool {
		summaBids += i.(*types.DepthItem).GetQuantity()
		return true
	})
	return
}

func TestUpdateAskAndBid(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ds.SetBids(bids)
	ask := ds.GetAsk(1.951)
	bid := ds.GetBid(1.951)
	summaAsks, summaBids := summaAsksAndBids(ds)
	assert.Equal(t, types.QuantityType(217.9), ask.(*types.DepthItem).GetQuantity())
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsksSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBidsSummaQuantity()), 6))
	ds.UpdateAsk(1.951, 300.0)
	ask = ds.GetAsk(1.951)
	bid = ds.GetBid(1.951)
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, types.QuantityType(300.0), ask.(*types.DepthItem).GetQuantity())
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsksSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBidsSummaQuantity()), 6))

	ds.UpdateBid(1.951, 300.0)
	ask = ds.GetAsk(1.951)
	bid = ds.GetBid(1.951)
	assert.Nil(t, ask)
	assert.Equal(t, types.QuantityType(300.0), bid.(*types.DepthItem).GetQuantity())
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsksSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBidsSummaQuantity()), 6))
	ds.SetBid(2.0, 100)
	assert.Equal(t, types.QuantityType(1771.3999999999999), ds.GetBidsSummaQuantity())
	ds.DeleteBid(2.0)
	assert.Equal(t, types.QuantityType(1671.3999999999999), ds.GetBidsSummaQuantity())
	ds.DeleteBid(2.0)
	assert.Equal(t, types.QuantityType(1671.3999999999999), ds.GetBidsSummaQuantity())
}

func TestGetFilteredByPercentAsksAndBids(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	ds.SetAsks(asks)
	normalizedAsks, _, _, _ := ds.GetFilteredByPercentAsks()
	normalizedBids, _, _, _ := ds.GetFilteredByPercentBids()
	assert.NotNil(t, normalizedAsks)
	assert.NotNil(t, normalizedBids)
	normalizedAsksArray := make([]types.DepthItem, 0)
	normalizedBidsArray := make([]types.DepthItem, 0)
	normalizedAsks.Ascend(func(i btree.Item) bool {
		normalizedAsksArray = append(normalizedAsksArray, *i.(*types.DepthItem))
		return true
	})
	normalizedBids.Ascend(func(i btree.Item) bool {
		normalizedBidsArray = append(normalizedBidsArray, *i.(*types.DepthItem))
		return true
	})
	assert.Equal(t, 8, len(normalizedAsksArray))
	assert.Equal(t, 8, len(normalizedBidsArray))
}

func TestDepthInterface(t *testing.T) {
	test := func(ds depth_interface.Depth) {
		ds.UpdateBid(1.93, 300.0)
		bid := ds.GetBid(1.93)
		assert.Equal(t, types.QuantityType(300.0), bid.(*types.DepthItem).GetQuantity())
		ds.UpdateAsk(1.951, 300.0)
		ask := ds.GetAsk(1.951)
		assert.Equal(t, types.QuantityType(300.0), ask.(*types.DepthItem).GetQuantity())
	}
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	ds.SetAsks(asks)
	test(ds)
}

func TestAsksAndBidMiddleQuantity(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
		initDepths(ds)
		asksMiddle := ds.GetAsksMiddleQuantity()
		assert.Equal(t, types.QuantityType(18.0), asksMiddle)
		bidsMiddle := ds.GetBidsMiddleQuantity()
		assert.Equal(t, types.QuantityType(18.0), bidsMiddle)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
		ds.SetAsks(asks)
		ds.SetBids(bids)
		asksMiddle := ds.GetAsksMiddleQuantity()
		assert.Equal(t, types.QuantityType(148.3375), asksMiddle)
		bidsMiddle := ds.GetBidsMiddleQuantity()
		assert.Equal(t, types.QuantityType(171.42499999999998), bidsMiddle)
	}()
}

func TestAsksAndBidStandardDeviation(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
		initDepths(ds)
		asksSquares := ds.GetAsksStandardDeviation()
		assert.Equal(t, 7.483314773547883, asksSquares)
		bidsSquares := ds.GetBidsStandardDeviation()
		assert.Equal(t, 7.483314773547883, bidsSquares)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depth_types.DepthStreamRate100ms)
		ds.SetAsks(asks)
		ds.SetBids(bids)
		asksSquares := ds.GetAsksStandardDeviation()
		assert.Equal(t, 39.70157230828522, asksSquares)
		bidsSquares := ds.GetBidsStandardDeviation()
		assert.Equal(t, 30.873805644915233, bidsSquares)
	}()
}

// func TestAddAskAndBidNormalized(t *testing.T) {
// 	func() {
// 		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
// 		ds.SetAsk(800, 100)
// 		ds.SetAsk(750, 150)
// 		askNorm1, _ := ds.GetNormalizedAsk(800)
// 		assert.Equal(t, types.PriceType(800.0), askNorm1.GetNormalizedPrice())
// 		assert.Equal(t, types.QuantityType(250.0), askNorm1.GetQuantity())
// 		ds.DeleteAsk(800)
// 		askNorm1, _ = ds.GetNormalizedAsk(800)
// 		assert.Equal(t, types.PriceType(800.0), askNorm1.GetNormalizedPrice())
// 		assert.Equal(t, types.QuantityType(150.0), askNorm1.GetQuantity())
// 	}()
// 	func() {
// 		asks, bids := getTestDepths()
// 		ds := depth_types.New(degree, "DOGEUSDT", true, 10, 100, -1, depth_types.DepthStreamRate100ms)
// 		ds.SetAsks(asks)
// 		ds.SetBids(bids)
// 		askNorm1, _ := ds.GetNormalizedAsk(1.953)
// 		assert.Equal(t, types.PriceType(2.0), askNorm1.GetNormalizedPrice())
// 		assert.Equal(t, types.QuantityType(1186.7), askNorm1.GetQuantity())
// 	}()
// }

// func TestGetNormalizedAsksAndBids(t *testing.T) {
// 	func() {
// 		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
// 		ds.SetAsk(800, 100)
// 		ds.SetAsk(750, 150)
// 		asksNorm := ds.GetNormalizedAsks()
// 		assert.Equal(t, 1, asksNorm.Len())
// 	}()
// 	func() {
// 		asks, bids := getTestDepths()
// 		ds := depth_types.New(degree, "DOGEUSDT", true, 10, 100, -1, depth_types.DepthStreamRate100ms)
// 		ds.SetAsks(asks)
// 		ds.SetBids(bids)
// 		asksNorm := ds.GetNormalizedAsks()
// 		assert.Equal(t, 1, asksNorm.Len())
// 	}()
// }

func TestAskAndBidDelete(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
		ds.SetAsk(800, 100)
		ds.SetAsk(750, 150)
		ds.DeleteAsk(800)
		ask := ds.GetAsk(800)
		assert.Nil(t, ask)
		assert.Equal(t, 1, ds.GetAsks().Len())
		ds.DeleteAsk(750)
		ask = ds.GetAsk(750)
		assert.Nil(t, ask)
		assert.Equal(t, 0, ds.GetAsks().Len())
	}()
}

func TestAskAndBidSummaQuantity(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
		ds.SetAsk(800, 100)
		assert.Equal(t, types.QuantityType(100.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(790, 100)
		assert.Equal(t, types.QuantityType(200.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(780, 100)
		assert.Equal(t, types.QuantityType(300.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(770, 100)
		assert.Equal(t, types.QuantityType(400.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(760, 100)
		assert.Equal(t, types.QuantityType(500.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(750, 100)
		assert.Equal(t, types.QuantityType(600.0), ds.GetAsksSummaQuantity())
		ds.SetAsk(740, 100)
		assert.Equal(t, types.QuantityType(700.0), ds.GetAsksSummaQuantity())
	}()
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
		ds.SetBid(800, 100)
		assert.Equal(t, types.QuantityType(100.0), ds.GetBidsSummaQuantity())
		ds.SetBid(790, 100)
		assert.Equal(t, types.QuantityType(200.0), ds.GetBidsSummaQuantity())
		ds.SetBid(780, 100)
		assert.Equal(t, types.QuantityType(300.0), ds.GetBidsSummaQuantity())
		ds.SetBid(770, 100)
		assert.Equal(t, types.QuantityType(400.0), ds.GetBidsSummaQuantity())
		ds.SetBid(760, 100)
		assert.Equal(t, types.QuantityType(500.0), ds.GetBidsSummaQuantity())
		ds.SetBid(750, 100)
		assert.Equal(t, types.QuantityType(600.0), ds.GetBidsSummaQuantity())
		ds.SetBid(740, 100)
		assert.Equal(t, types.QuantityType(700.0), ds.GetBidsSummaQuantity())
	}()
}

// func TestAskAndSummaNormalizedQuantity(t *testing.T) {
// 	ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
// 	testSetAsk := func(price types.PriceType, quantity types.QuantityType) (summa, summaTest, summaTest2 types.QuantityType) {
// 		ds.SetAsk(price, quantity)
// 		askN, _ := ds.GetNormalizedAsk(price)
// 		summa = askN.GetQuantity()
// 		summaTest, summaTest2 = ds.GetNormalizedAskSumma(price)
// 		return
// 	}
// 	summa, summaTest, summaTest2 := testSetAsk(800, 100)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	summa, summaTest, summaTest2 = testSetAsk(750, 100)
// 	assert.Equal(t, types.QuantityType(200), summa)
// 	assert.Equal(t, types.QuantityType(200), summaTest)
// 	assert.Equal(t, types.QuantityType(200), summaTest2)
// 	summa, summaTest, summaTest2 = testSetAsk(700, 100)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	summa, summaTest, summaTest2 = testSetAsk(650, 100)
// 	assert.Equal(t, types.QuantityType(200), summa)
// 	assert.Equal(t, types.QuantityType(200), summaTest)
// 	assert.Equal(t, types.QuantityType(200), summaTest2)
// 	summa, summaTest, summaTest2 = testSetAsk(600, 100)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	summa, summaTest, summaTest2 = testSetAsk(550, 100)
// 	assert.Equal(t, types.QuantityType(200), summa)
// 	assert.Equal(t, types.QuantityType(200), summaTest)
// 	assert.Equal(t, types.QuantityType(200), summaTest2)

// 	summaAsks1 := ds.GetAsksSummaQuantity()
// 	summaAsks2 := types.QuantityType(0)
// 	ds.GetAsks().Ascend(func(i btree.Item) bool {
// 		summaAsks2 += i.(*types.DepthItem).GetQuantity()
// 		return true
// 	})
// 	assert.Equal(t, summaAsks1, summaAsks2)
// 	summaAsks3 := types.QuantityType(0)
// 	ds.GetNormalizedAsks().Ascend(func(i btree.Item) bool {
// 		summaAsks3 += i.(*types.NormalizedItem).GetQuantity()
// 		return true
// 	})
// 	assert.Equal(t, summaAsks1, summaAsks3)
// 	assert.Equal(t, summaAsks2, summaAsks1)

// 	testDeleteAsk := func(price types.PriceType) (askN *types.NormalizedItem, summa, summaTest, summaTest2 types.QuantityType) {
// 		ds.DeleteAsk(price)
// 		askN, _ = ds.GetNormalizedAsk(price)
// 		summaTest, summaTest2 = ds.GetNormalizedAskSumma(price)
// 		if askN != nil {
// 			summa = askN.GetQuantity()
// 		}
// 		return
// 	}
// 	askN, summa, summaTest, summaTest2 := testDeleteAsk(800)
// 	assert.NotNil(t, askN)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	askN, summa, summaTest, summaTest2 = testDeleteAsk(750)
// 	assert.Nil(t, askN)
// 	assert.Equal(t, types.QuantityType(0), summa)
// 	assert.Equal(t, types.QuantityType(0), summaTest)
// 	assert.Equal(t, types.QuantityType(0), summaTest2)
// 	askN, summa, summaTest, summaTest2 = testDeleteAsk(700)
// 	assert.NotNil(t, askN)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	askN, summa, summaTest, summaTest2 = testDeleteAsk(650)
// 	assert.Nil(t, askN)
// 	assert.Equal(t, types.QuantityType(0), summa)
// 	assert.Equal(t, types.QuantityType(0), summaTest)
// 	assert.Equal(t, types.QuantityType(0), summaTest2)
// 	askN, summa, summaTest, summaTest2 = testDeleteAsk(600)
// 	assert.NotNil(t, askN)
// 	assert.Equal(t, types.QuantityType(100), summa)
// 	assert.Equal(t, types.QuantityType(100), summaTest)
// 	assert.Equal(t, types.QuantityType(100), summaTest2)
// 	askN, summa, summaTest, summaTest2 = testDeleteAsk(550)
// 	assert.Nil(t, askN)
// 	assert.Equal(t, types.QuantityType(0), summa)
// 	assert.Equal(t, types.QuantityType(0), summaTest)
// 	assert.Equal(t, types.QuantityType(0), summaTest2)

// 	summaAsks1 = ds.GetAsksSummaQuantity()
// 	summaAsks2 = types.QuantityType(0)
// 	ds.GetAsks().Ascend(func(i btree.Item) bool {
// 		summaAsks2 += i.(*types.DepthItem).GetQuantity()
// 		return true
// 	})
// 	assert.Equal(t, summaAsks1, summaAsks2)
// 	summaAsks3 = types.QuantityType(0)
// 	ds.GetNormalizedAsks().Ascend(func(i btree.Item) bool {
// 		summaAsks3 += i.(*types.NormalizedItem).GetQuantity()
// 		return true
// 	})
// 	assert.Equal(t, summaAsks1, summaAsks3)
// 	assert.Equal(t, summaAsks2, summaAsks1)
// }

// func TestAskAndBidMinMaxQuantity(t *testing.T) {
// 	func() {
// 		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depth_types.DepthStreamRate100ms)
// 		minTest := func() types.QuantityType {
// 			min, err := ds.AskMin()
// 			if err != nil {
// 				return 0
// 			}
// 			return min.GetQuantity()
// 		}
// 		maxTest := func() types.QuantityType {
// 			max, err := ds.AskMax()
// 			if err != nil {
// 				return 0
// 			}
// 			return max.GetQuantity()
// 		}
// 		ds.SetAsk(800, 100)
// 		assert.Equal(t, types.QuantityType(100.0), minTest())
// 		assert.Equal(t, types.QuantityType(100.0), maxTest())
// 		ds.SetAsk(790, 200)
// 		assert.Equal(t, types.QuantityType(100.0), minTest())
// 		assert.Equal(t, types.QuantityType(200.0), maxTest())
// 		ds.SetAsk(780, 300)
// 		assert.Equal(t, types.QuantityType(100.0), minTest())
// 		assert.Equal(t, types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(800)
// 		assert.Equal(t, types.QuantityType(200.0), minTest())
// 		assert.Equal(t, types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(790)
// 		assert.Equal(t, types.QuantityType(300.0), minTest())
// 		assert.Equal(t, types.QuantityType(300.0), maxTest())
// 		ds.DeleteAsk(780)
// 		assert.Equal(t, types.QuantityType(0.0), minTest())
// 		assert.Equal(t, types.QuantityType(0.0), maxTest())
// 	}()
// }
