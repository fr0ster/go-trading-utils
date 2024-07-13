package depth_test

import (
	"testing"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestLockUnlock(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	d.Lock()
	defer d.Unlock()

	// Add assertions here to verify that the lock and unlock operations are working correctly
	assert.True(t, true)
}

func TestSetAndGetAsk(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	price := 100.0
	quantity := 10.0

	d.SetAsk(price, quantity)
	ask := d.GetAsk(price)

	// Add assertions here to verify that the ask is set and retrieved correctly
	assert.NotNil(t, ask)
	assert.Equal(t, price, ask.(*depth_types.DepthItem).Price)
}

func TestSetAndGetBid(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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

func TestGetTargetAsksBidPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	asks, bids, summaAsks, summaBids := d.GetTargetAsksBidPrice(d.GetAsksSummaQuantity()*0.2, d.GetBidsSummaQuantity()*0.2)
	assert.NotNil(t, asks)
	assert.NotNil(t, bids)
	assert.Equal(t, 500.0, asks.Price)
	assert.Equal(t, 10.0, summaAsks)
	assert.Equal(t, 400.0, bids.Price)
	assert.Equal(t, 10.0, summaBids)
}

func TestGetAsksAndBidSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, 80.0, d.GetAsksSummaQuantity())
	assert.Equal(t, 80.0, d.GetBidsSummaQuantity())
	summaAsks := d.GetAsksSumma(650.0)
	assert.Equal(t, 40.0, summaAsks)
	summaBids := d.GetBidsSumma(250.0)
	assert.Equal(t, 20.0, summaBids)
}

func TestGetAsksMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly
	limit := d.GetAsksMaxUpToPrice(650.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 600.0, limit.Price)
	assert.Equal(t, 30.0, limit.Quantity)
}

func TestGetBidsMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly when no price is provided
	limit := d.GetBidsMaxDownToPrice(250)
	assert.NotNil(t, limit)
	assert.Equal(t, 400.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetAsksMaxUpToSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToSumma method works correctly
	limit := d.GetAsksMaxUpToSumma(35.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 500.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetBidsMaxDownToSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetBidsMaxDownToSumma method works correctly
	limit := d.GetBidsMaxDownToSumma(35.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 400.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetFilteredByPercentAsks(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetSummaOfAsksFromRange method works correctly
	summaAsk := d.GetSummaOfAsksFromRange(600.0, 800.0, func(d *depth_types.DepthItem) bool { return true })
	assert.Equal(t, 60.0, summaAsk)
	summaBid := d.GetSummaOfBidsFromRange(300.0, 50.0, func(d *depth_types.DepthItem) bool { return true })
	assert.Equal(t, 70.0, summaBid)
}

func TestMinMax(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
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
