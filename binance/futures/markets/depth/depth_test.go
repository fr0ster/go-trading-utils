package depth_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	"github.com/fr0ster/go-trading-utils/utils"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

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

func TestGetDepthNew(t *testing.T) {
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	testDepthTree := depth_types.New(3, "SUSHIUSDT")
	// Add additional assertions if needed
	assert.NotEmpty(t, testDepthTree)
}

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	futures := futures.NewClient(api_key, secret_key)

	// Add more test cases here
	testDepthTree := depth_types.New(3, "SUSHIUSDT")
	err := futures_depth.Init(testDepthTree, futures, 10)
	assert.NoError(t, err)
}

func TestGetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
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
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	bid := ds.GetBid(1.93)
	if bid == nil {
		t.Errorf("Failed to get bid")
	}
}

func TestSetAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetAsks(asks)
	ask := depth_types.DepthItem{Price: 1.96, Quantity: 200.0}
	ds.SetAsk(ask.Price, ask.Quantity)
	assert.NotNil(t, 1.96, ds.GetAsk(1.96))
}

func TestSetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	bid := depth_types.DepthItem{Price: 1.96, Quantity: 200.0}
	ds.SetBid(bid.Price, bid.Quantity)
	assert.NotNil(t, 1.96, ds.GetBid(1.96))
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
	ds := depth_types.New(3, "SUSHIUSDT")
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

func TestGetNormalizedDepth(t *testing.T) {
	asks, bids := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	ds.SetAsks(asks)
	normalizedAsks := ds.GetNormalizedAsks()
	normalizedBids := ds.GetNormalizedBids()
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
	_, bids := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	test(ds)
}
