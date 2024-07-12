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
	depth.SetAsk(600.0, 30.0)
	depth.SetAsk(500.0, 20.0)
	depth.SetAsk(400.0, 10.0)
	depth.SetBid(300.0, 10.0)
	depth.SetBid(200.0, 20.0)
	depth.SetBid(100.0, 30.0)
}

func TestGetTargetAsksBidPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetTargetAsksBidPrice method works correctly
	asks, bids, summaAsks, summaBids := d.GetTargetAsksBidPrice(d.GetAsksSummaQuantity()*0.2, d.GetBidsSummaQuantity()*0.2)
	assert.NotNil(t, asks)
	assert.NotNil(t, bids)
	assert.Equal(t, 400.0, asks.Price)
	assert.Equal(t, 10.0, summaAsks)
	assert.Equal(t, 300.0, bids.Price)
	assert.Equal(t, 10.0, summaBids)
}

func TestGetAsksAndBidSumma(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksSummaQuantity and GetBidsSummaQuantity methods work correctly
	assert.Equal(t, 60.0, d.GetAsksSummaQuantity())
	assert.Equal(t, 60.0, d.GetBidsSummaQuantity())
	summaAsks := d.GetAsksSumma(450.0)
	assert.Equal(t, 10.0, summaAsks)
	summaBids := d.GetBidsSumma(250.0)
	assert.Equal(t, 10.0, summaBids)
}

func TestGetAsksMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly
	limit := d.GetAsksMaxUpToPrice(450.0)
	assert.NotNil(t, limit)
	assert.Equal(t, 400.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}

func TestGetBidsMaxUpToPrice(t *testing.T) {
	d := depth_types.New(degree, "BTCUSDT", depth_types.DepthAPILimit20, depth_types.DepthStreamRate100ms)
	initDepths(d)
	// Add assertions here to verify that the GetAsksMaxUpToPrice method works correctly when no price is provided
	limit := d.GetBidsMaxDownToPrice(250)
	assert.NotNil(t, limit)
	assert.Equal(t, 300.0, limit.Price)
	assert.Equal(t, 10.0, limit.Quantity)
}
