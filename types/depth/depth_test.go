package depth_test

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"

	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	degree = 3
)

func TestLockUnlock(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	d.Lock()
	defer d.Unlock()

	// Add assertions here to verify that the lock and unlock operations are working correctly
	assert.True(t, true)
}

func TestSetAndGetAsk(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	price := 100.0
	quantity := 10.0

	d.SetAsk(price, quantity)
	ask := d.GetAsk(price)

	// Add assertions here to verify that the ask is set and retrieved correctly
	assert.NotNil(t, ask)
	assert.Equal(t, price, ask.(*depth_types.DepthItem).Price)
}

func TestSetAndGetBid(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	price := 200.0
	quantity := 20.0

	d.SetBid(price, quantity)
	bid := d.GetBid(price)

	// Add assertions here to verify that the bid is set and retrieved correctly
	assert.NotNil(t, bid)
	assert.Equal(t, price, bid.(*depth_types.DepthItem).Price)
}

// Add more tests here based on your requirements

func initDepths(depth *depth_types.Depth) {
	depth.SetAsk(1000.0, 10.0)
	depth.SetAsk(800.0, 10.0)
	depth.SetAsk(700.0, 20.0)
	depth.SetAsk(600.0, 30.0)
	depth.SetAsk(500.0, 10.0)
	depth.SetBid(400.0, 10.0)
	depth.SetBid(300.0, 10.0)
	depth.SetBid(200.0, 20.0)
	depth.SetBid(100.0, 30.0)
	depth.SetBid(50.0, 10.0)
}

func getTestDepths() (asks *btree.BTree, bids *btree.BTree) {
	bids = btree.New(3)
	bidList := []depth_types.DepthItem{
		{Price: 1.92, Quantity: 150.2},
		{Price: 1.93, Quantity: 155.4}, // local maxima
		{Price: 1.94, Quantity: 150.0},
		{Price: 1.941, Quantity: 130.4},
		{Price: 1.947, Quantity: 172.1},
		{Price: 1.948, Quantity: 187.4},
		{Price: 1.949, Quantity: 236.1}, // local maxima
		{Price: 1.95, Quantity: 189.8},
	}
	asks = btree.New(3)
	askList := []depth_types.DepthItem{
		{Price: 1.951, Quantity: 217.9}, // local maxima
		{Price: 1.952, Quantity: 179.4},
		{Price: 1.953, Quantity: 180.9}, // local maxima
		{Price: 1.954, Quantity: 148.5},
		{Price: 1.955, Quantity: 120.0},
		{Price: 1.956, Quantity: 110.0},
		{Price: 1.957, Quantity: 140.0}, // local maxima
		{Price: 1.958, Quantity: 90.0},
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
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	func() {
		asks, bids, summaAsks, summaBids := d.GetAsksBidFirstMaxAndSumma(d.GetAsksSummaQuantity()*0.3, d.GetBidsSummaQuantity()*0.3)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, 500.0, asks.Price)
		assert.Equal(t, 10.0, summaAsks)
		assert.Equal(t, 300.0, bids.Price)
		assert.Equal(t, 20.0, summaBids)
	}()
	func() {
		asks, bids, summaAsks, summaBids := d.GetAsksBidFirstMaxAndSumma(d.GetAsksSummaQuantity()*0.3, d.GetBidsSummaQuantity()*0.3, true)
		assert.NotNil(t, asks)
		assert.NotNil(t, bids)
		assert.Equal(t, 500.0, asks.Price)
		assert.Equal(t, 10.0, summaAsks)
		assert.Equal(t, 400.0, bids.Price)
		assert.Equal(t, 10.0, summaBids)
	}()
}

func TestGetAsksAndBidSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, 80.0, d.GetAsksSummaQuantity())
	assert.Equal(t, 80.0, d.GetBidsSummaQuantity())
	maxAsks, summaAsks := d.GetAsksMaxAndSummaUp(650.0)
	assert.Equal(t, 40.0, summaAsks)
	assert.Equal(t, 600.0, maxAsks.Price)
	maxBids, summaBids := d.GetBidsMaxAndSummaDown(250.0)
	assert.Equal(t, 20.0, summaBids)
	assert.Equal(t, 400.0, maxBids.Price)
}

func TestGetAsksMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly
	limit, _ := d.GetAsksMaxAndSummaUp(650.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 600.0, limit.Price)
	assert.Equal(t, 30.0, limit.Quantity)
}

func TestGetBidsMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly when no price is provided
	limit, _ := d.GetBidsMaxAndSummaDown(250)
	assert.NotNil(t, limit)
	assert.Equal(t, 400.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetAsksMaxUpToSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToSumma method works correctly
	limit := d.GetAsksMaxUpToSumma(35.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 500.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetBidsMaxDownToSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetBidsMaxDownToSumma method works correctly
	limit := d.GetBidsMaxDownToSumma(35.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 400.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetFilteredByPercentAsks(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentAsks method works correctly
	filtered, summa, max, min := d.GetFilteredByPercentAsks(func(i *depth_types.DepthItem) bool {
		return i.Quantity*100/d.GetAsksSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, 30.0, summa)
	assert.Equal(t, 30.0, max)
	assert.Equal(t, 30.0, min)

	filtered, summa, max, min = d.GetFilteredByPercentAsks(func(i *depth_types.DepthItem) bool {
		return i.Quantity*100/d.GetAsksSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, 50.0, summa)
	assert.Equal(t, 20.0, max)
	assert.Equal(t, 10.0, min)
}

func TestGetFilteredByPercentBids(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetFilteredByPercentBids method works correctly
	filtered, summa, max, min := d.GetFilteredByPercentBids(func(i *depth_types.DepthItem) bool {
		return i.Quantity*100/d.GetBidsSummaQuantity() > 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 1, filtered.Len())
	assert.Equal(t, 30.0, summa)
	assert.Equal(t, 30.0, max)
	assert.Equal(t, 30.0, min)

	filtered, summa, max, min = d.GetFilteredByPercentBids(func(i *depth_types.DepthItem) bool {
		return i.Quantity*100/d.GetBidsSummaQuantity() < 30
	})
	assert.NotNil(t, filtered)
	assert.Equal(t, 4, filtered.Len())
	assert.Equal(t, 50.0, summa)
	assert.Equal(t, 20.0, max)
	assert.Equal(t, 10.0, min)
}

func TestGetSummaOfAsksAndBidFromRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetSummaOfAsksFromRange method works correctly
	summaAsk, max, min := d.GetSummaOfAsksFromRange(600.0, 800.0, func(d *depth_types.DepthItem) bool { return true })
	assert.Equal(t, 30.0, summaAsk)
	assert.Equal(t, 20.0, max)
	assert.Equal(t, 10.0, min)
	summaBid, max, min := d.GetSummaOfBidsFromRange(300.0, 50.0, func(d *depth_types.DepthItem) bool { return true })
	assert.Equal(t, 60.0, summaBid)
	assert.Equal(t, 30.0, max)
	assert.Equal(t, 10.0, min)
}

func TestMinMax(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the Min and Max methods work correctly
	min, err := d.AskMin()
	assert.Nil(t, err)
	assert.Equal(t, 500.0, min.Price)
	max, err := d.AskMax()
	assert.Nil(t, err)
	assert.Equal(t, 600.0, max.Price)
	min, err = d.BidMin()
	assert.Nil(t, err)
	assert.Equal(t, 400.0, min.Price)
	max, err = d.BidMax()
	assert.Nil(t, err)
	assert.Equal(t, 100.0, max.Price)
}

func TestGetAsksAndBidSummaAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, 80.0, d.GetAsksSummaQuantity())
	assert.Equal(t, 80.0, d.GetBidsSummaQuantity())
	maxAsk1, summaAsks1 := d.GetAsksMaxAndSummaUp(600.0)
	assert.Equal(t, 40.0, summaAsks1)
	assert.Equal(t, 600.0, maxAsk1.Price)
	assert.Equal(t, 30.0, maxAsk1.Quantity)
	summaAsks2, max, min := d.GetSummaOfAsksFromRange(600.0, 800.0)
	assert.Equal(t, 30.0, summaAsks2)
	assert.Equal(t, 20.0, max)
	assert.Equal(t, 10.0, min)
	maxAsk3, summaAsks3 := d.GetAsksMaxAndSummaUp(800.0)
	assert.Equal(t, 70.0, summaAsks3)
	assert.Equal(t, 600.0, maxAsk3.Price)
	assert.Equal(t, 30.0, maxAsk3.Quantity)
	assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
	maxBid1, summaBids1 := d.GetBidsMaxAndSummaDown(300.0)
	assert.Equal(t, 20.0, summaBids1)
	assert.Equal(t, 400.0, maxBid1.Price)
	assert.Equal(t, 10.0, maxBid1.Quantity)
	summaBids2, max, min := d.GetSummaOfBidsFromRange(300.0, 50.0)
	assert.Equal(t, 60.0, summaBids2)
	assert.Equal(t, 30.0, max)
	assert.Equal(t, 10.0, min)
	maxBid3, summaBids3 := d.GetBidsMaxAndSummaDown(50.0)
	assert.Equal(t, 80.0, summaBids3)
	assert.Equal(t, 100.0, maxBid3.Price)
	assert.Equal(t, 30.0, maxBid3.Quantity)
	assert.Equal(t, summaBids2, summaBids3-summaBids1)
}

func TestGetTargetAsksBidPriceAndRange(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, 80.0, d.GetAsksSummaQuantity())
	assert.Equal(t, 80.0, d.GetBidsSummaQuantity())
	ask1, bid1, summaAsks1, summaBids1 := d.GetAsksBidFirstMaxAndSumma(20, 20)
	ask2, bid2, summaAsks3, summaBids3 := d.GetAsksBidFirstMaxAndSumma(50, 50)
	assert.Equal(t, 10.0, summaAsks1)
	assert.Equal(t, 10.0, summaBids1)
	assert.Equal(t, 40.0, summaAsks3)
	assert.Equal(t, 40.0, summaBids3)
	summaAsks2, max, min := d.GetSummaOfAsksFromRange(ask1.Price, ask2.Price)
	assert.Equal(t, 30.0, max)
	assert.Equal(t, 30.0, min)
	summaBids2, max, min := d.GetSummaOfBidsFromRange(bid1.Price, bid2.Price)
	assert.Equal(t, 20.0, max)
	assert.Equal(t, 10.0, min)
	assert.Equal(t, summaAsks2, summaAsks3-summaAsks1)
	assert.Equal(t, summaBids2, summaBids3-summaBids1)
}

func TestGetTargetPrices(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", false, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetPrices method works correctly
	ask1, bid1 := d.GetTargetPrices(20)
	ask2, bid2 := d.GetTargetPrices(50)
	assert.Equal(t, 500.0, ask1)
	assert.Equal(t, 400.0, bid1)
	assert.Equal(t, 500.0, ask2)
	assert.Equal(t, 400.0, bid2)
}

func TestGetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	bid := ds.GetBid(1.93)
	if bid == nil {
		t.Errorf("Failed to get bid")
	}
}

func TestSetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ask := depth_types.DepthItem{Price: 1.96, Quantity: 200.0}
	ds.SetAsk(ask.Price, ask.Quantity)
	if ds.GetAsk(1.96) == nil {
		t.Errorf("Failed to set ask")
	}
}

func TestSetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	bid := depth_types.DepthItem{Price: 1.96, Quantity: 200.0}
	ds.SetBid(bid.Price, bid.Quantity)
	if ds.GetBid(1.96) == nil {
		t.Errorf("Failed to set bid")
	}
}

func TestRestrictAskAndBidDown(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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

func summaAsksAndBids(ds *depth_types.Depth) (summaAsks, summaBids float64) {
	ds.GetAsks().Ascend(func(i btree.Item) bool {
		summaAsks += i.(*depth_types.DepthItem).Quantity
		return true
	})
	ds.GetBids().Ascend(func(i btree.Item) bool {
		summaBids += i.(*depth_types.DepthItem).Quantity
		return true
	})
	return
}

func TestUpdateAskAndBid(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetAsks(asks)
	ds.SetBids(bids)
	ask := ds.GetAsk(1.951)
	bid := ds.GetBid(1.951)
	summaAsks, summaBids := summaAsksAndBids(ds)
	assert.Equal(t, 217.9, ask.(*depth_types.DepthItem).Quantity)
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(summaAsks, 6), utils.RoundToDecimalPlace(ds.GetAsksSummaQuantity(), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(summaBids, 6), utils.RoundToDecimalPlace(ds.GetBidsSummaQuantity(), 6))
	ds.UpdateAsk(1.951, 300.0)
	ask = ds.GetAsk(1.951)
	bid = ds.GetBid(1.951)
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, 300.0, ask.(*depth_types.DepthItem).Quantity)
	assert.Nil(t, bid)
	assert.Equal(t, utils.RoundToDecimalPlace(summaAsks, 6), utils.RoundToDecimalPlace(ds.GetAsksSummaQuantity(), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(summaBids, 6), utils.RoundToDecimalPlace(ds.GetBidsSummaQuantity(), 6))

	ds.UpdateBid(1.951, 300.0)
	ask = ds.GetAsk(1.951)
	bid = ds.GetBid(1.951)
	assert.Nil(t, ask)
	assert.Equal(t, 300.0, bid.(*depth_types.DepthItem).Quantity)
	summaAsks, summaBids = summaAsksAndBids(ds)
	assert.Equal(t, utils.RoundToDecimalPlace(summaAsks, 6), utils.RoundToDecimalPlace(ds.GetAsksSummaQuantity(), 6))
	assert.Equal(t, utils.RoundToDecimalPlace(summaBids, 6), utils.RoundToDecimalPlace(ds.GetBidsSummaQuantity(), 6))
}

func TestGetFilteredByPercentAsksAndBids(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	ds.SetAsks(asks)
	normalizedAsks, _, _, _ := ds.GetFilteredByPercentAsks()
	normalizedBids, _, _, _ := ds.GetFilteredByPercentBids()
	assert.NotNil(t, normalizedAsks)
	assert.NotNil(t, normalizedBids)
	normalizedAsksArray := make([]depth_types.DepthItem, 0)
	normalizedBidsArray := make([]depth_types.DepthItem, 0)
	normalizedAsks.Ascend(func(i btree.Item) bool {
		normalizedAsksArray = append(normalizedAsksArray, *i.(*depth_types.DepthItem))
		return true
	})
	normalizedBids.Ascend(func(i btree.Item) bool {
		normalizedBidsArray = append(normalizedBidsArray, *i.(*depth_types.DepthItem))
		return true
	})
	assert.Equal(t, 8, len(normalizedAsksArray))
	assert.Equal(t, 8, len(normalizedBidsArray))
}

func TestDepthInterface(t *testing.T) {
	test := func(ds depth_interface.Depth) {
		ds.UpdateBid(1.93, 300.0)
		bid := ds.GetBid(1.93)
		assert.Equal(t, 300.0, bid.(*depth_types.DepthItem).Quantity)
		ds.UpdateAsk(1.951, 300.0)
		ask := ds.GetAsk(1.951)
		assert.Equal(t, 300.0, ask.(*depth_types.DepthItem).Quantity)
	}
	asks, bids := getTestDepths()
	ds := depth_types.New(degree, "BTCUSDT", true, 10, 75, depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	ds.SetBids(bids)
	ds.SetAsks(asks)
	test(ds)
}
