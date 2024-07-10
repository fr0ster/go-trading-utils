package depth_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"

	spot_depth "github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"

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
	if testDepthTree == nil {
		t.Errorf("GetDepthTree returned an empty map")
	}

	// Add additional assertions if needed
}

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	spot := binance.NewClient(api_key, secret_key)

	// Add more test cases here
	testDepthTree := depth_types.New(3, "SUSHIUSDT")
	err := spot_depth.Init(testDepthTree, spot, 10)
	if err != nil {
		t.Errorf("Failed to initialize depth tree: %v", err)
	}
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
	if ds.GetAsk(1.96) == nil {
		t.Errorf("Failed to set ask")
	}
}

func TestSetBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	bid := depth_types.DepthItem{Price: 1.96, Quantity: 200.0}
	ds.SetBid(bid.Price, bid.Quantity)
	if ds.GetBid(1.96) == nil {
		t.Errorf("Failed to set bid")
	}
}

func TestUpdateAsk(t *testing.T) {
	asks, _ := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetAsks(asks)
	ds.UpdateAsk(1.951, 300.0)
	ask := ds.GetAsk(1.951)
	assert.Equal(t, 300.0, ask.(*depth_types.DepthItem).Quantity)
	ds.UpdateAsk(1.000, 100.0)
	assert.NotNil(t, ds.GetAsk(1.000))
}

func TestUpdateBid(t *testing.T) {
	_, bids := getTestDepths()
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	ds.UpdateBid(1.93, 300.0)
	bid := ds.GetBid(1.93)
	assert.Equal(t, 300.0, bid.(*depth_types.DepthItem).Quantity)
	ds.UpdateBid(1.000, 100.0)
	assert.NotNil(t, ds.GetBid(1.000))
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
	ds := depth_types.New(3, "SUSHIUSDT")
	ds.SetBids(bids)
	ds.SetAsks(asks)
	test(ds)
}
