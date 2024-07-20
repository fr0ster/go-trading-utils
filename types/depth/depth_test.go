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
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetSummaQuantity()), 6))
	ds.UpdateAsk(items_types.NewAsk(1.951, 300.0))
	ask = ds.GetAsks().Get(items_types.NewAsk(1.951))
	bid = ds.GetBids().Get(items_types.NewBid(1.951))
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, items_types.QuantityType(300.0), ask.GetDepthItem().GetQuantity())
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetSummaQuantity()), 6))

	ds.UpdateBid(items_types.NewBid(1.951, 300.0))
	ask = ds.GetAsks().Get(items_types.NewAsk(1.951))
	bid = ds.GetBids().Get(items_types.NewBid(1.951))
	assert.Nil(t, ask)
	assert.Equal(t, items_types.QuantityType(300.0), bid.GetDepthItem().GetQuantity())
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaAsks), 6), utils.RoundToDecimalPlace(float64(ds.GetAsks().GetSummaQuantity()), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(float64(summaBids), 6), utils.RoundToDecimalPlace(float64(ds.GetBids().GetSummaQuantity()), 6))
	ds.GetBids().Set(items_types.NewBid(2.0, 100))
	assert.Equal(t, items_types.QuantityType(1771.3999999999999), ds.GetBids().GetSummaQuantity())
	ds.GetBids().Delete(items_types.NewBid(2.0))
	assert.Equal(t, items_types.QuantityType(1671.3999999999999), ds.GetBids().GetSummaQuantity())
	ds.GetBids().Delete(items_types.NewBid(2.0))
	assert.Equal(t, items_types.QuantityType(1671.3999999999999), ds.GetBids().GetSummaQuantity())
}

func TestAsksAndBidMiddleQuantity(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		initDepths(ds)
		asksMiddle := ds.GetAsks().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(18.0), asksMiddle)
		bidsMiddle := ds.GetBids().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(18.0), bidsMiddle)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().SetTree(asks)
		ds.GetBids().SetTree(bids)
		asksMiddle := ds.GetAsks().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(148.3375), asksMiddle)
		bidsMiddle := ds.GetBids().GetMiddleQuantity()
		assert.Equal(t, items_types.QuantityType(171.42499999999998), bidsMiddle)
	}()
}

func TestAsksAndBidStandardDeviation(t *testing.T) {
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		initDepths(ds)
		asksSquares := ds.GetAsks().GetStandardDeviation()
		assert.Equal(t, 7.483314773547883, asksSquares)
		bidsSquares := ds.GetBids().GetStandardDeviation()
		assert.Equal(t, 7.483314773547883, bidsSquares)
	}()
	func() {
		asks, bids := getTestDepths()
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, 2, depths_types.DepthStreamRate100ms)
		ds.GetAsks().SetTree(asks)
		ds.GetBids().SetTree(bids)
		asksSquares := ds.GetAsks().GetStandardDeviation()
		assert.Equal(t, 39.70157230828522, asksSquares)
		bidsSquares := ds.GetBids().GetStandardDeviation()
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
		assert.Equal(t, items_types.QuantityType(100.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(790, 100))
		assert.Equal(t, items_types.QuantityType(200.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(780, 100))
		assert.Equal(t, items_types.QuantityType(300.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(770, 100))
		assert.Equal(t, items_types.QuantityType(400.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(760, 100))
		assert.Equal(t, items_types.QuantityType(500.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(750, 100))
		assert.Equal(t, items_types.QuantityType(600.0), ds.GetAsks().GetSummaQuantity())
		ds.GetAsks().Set(items_types.NewAsk(740, 100))
		assert.Equal(t, items_types.QuantityType(700.0), ds.GetAsks().GetSummaQuantity())
	}()
	func() {
		ds := depth_types.New(degree, "BTCUSDT", true, 10, 100, 2, depths_types.DepthStreamRate100ms)
		ds.GetBids().Set(items_types.NewBid(800, 100))
		assert.Equal(t, items_types.QuantityType(100.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(790, 100))
		assert.Equal(t, items_types.QuantityType(200.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(780, 100))
		assert.Equal(t, items_types.QuantityType(300.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(770, 100))
		assert.Equal(t, items_types.QuantityType(400.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(760, 100))
		assert.Equal(t, items_types.QuantityType(500.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(750, 100))
		assert.Equal(t, items_types.QuantityType(600.0), ds.GetBids().GetSummaQuantity())
		ds.GetBids().Set(items_types.NewBid(740, 100))
		assert.Equal(t, items_types.QuantityType(700.0), ds.GetBids().GetSummaQuantity())
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
