package depth_test

import (
	"os"
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/markets/depth"
	depth_interface "github.com/fr0ster/go-trading-utils/interfaces/depth"
	"github.com/google/btree"
)

func getTestDepths() *depth.DepthBTree {
	testDepthTree := depth.DepthNew(3)
	records := []depth_interface.DepthItemType{
		{Price: 1.92, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.2},
		{Price: 1.93, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 155.4}, // local maxima
		{Price: 1.94, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 150.0},
		{Price: 1.941, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 130.4},
		{Price: 1.947, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 172.1},
		{Price: 1.948, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 187.4},
		{Price: 1.949, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 236.1}, // local maxima
		{Price: 1.95, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 189.8},
		{Price: 1.951, AskLastUpdateID: 2369068, AskQuantity: 217.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.952, AskLastUpdateID: 2369068, AskQuantity: 179.4, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.953, AskLastUpdateID: 2369068, AskQuantity: 180.9, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.954, AskLastUpdateID: 2369068, AskQuantity: 148.5, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.955, AskLastUpdateID: 2369068, AskQuantity: 120.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.956, AskLastUpdateID: 2369068, AskQuantity: 110.0, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.957, AskLastUpdateID: 2369068, AskQuantity: 140.0, BidLastUpdateID: 0, BidQuantity: 0}, // local maxima
		{Price: 1.958, AskLastUpdateID: 2369068, AskQuantity: 90.0, BidLastUpdateID: 0, BidQuantity: 0},
	}
	for _, record := range records {
		testDepthTree.ReplaceOrInsert(&record)
	}

	return testDepthTree
}

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := true

	// Add more test cases here
	testDepthTree := depth.DepthNew(3)
	err := testDepthTree.Init(api_key, secret_key, "BTCUSDT", UseTestnet)
	if err != nil {
		t.Errorf("Failed to initialize depth tree: %v", err)
	}
}

func TestGetDepthNew(t *testing.T) {
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	testDepthTree := depth.DepthNew(3)
	if testDepthTree == nil {
		t.Errorf("GetDepthTree returned an empty map")
	}

	// Add additional assertions if needed
}

func TestGetDepthMaxBidMinAsk(t *testing.T) {
	testDepthTree := getTestDepths()
	// Call the function being tested
	maxBid, minAsk := testDepthTree.GetMaxBidMinAsk()
	if maxBid.Price != 1.95 {
		t.Errorf("GetDepthMaxBid returned an incorrect max bid price")
	}
	if minAsk.Price != 1.951 {
		t.Errorf("GetDepthMinAsk returned an incorrect min ask price")
	}
}

func TestGetDepthMaxBidQtyMaxAskQty(t *testing.T) {
	testDepthTree := getTestDepths()
	// Call the function being tested
	maxBid, minAsk := testDepthTree.GetMaxBidQtyMaxAskQty()
	if maxBid.Price != 1.949 {
		t.Errorf("GetDepthMaxBid returned an incorrect max bid price")
	}
	if minAsk.Price != 1.951 {
		t.Errorf("GetDepthMinAsk returned an incorrect min ask price")
	}
}

func TestGetDepthBidLocalMaxima(t *testing.T) {
	testDepthTree := getTestDepths()

	bidLocalsMaxima := testDepthTree.GetBidQtyLocalMaxima()
	askLocalMaxima := testDepthTree.GetAskQtyLocalMaxima()

	// Add assertions to check the correctness of the returned map
	if bidLocalsMaxima.Get(&depth_interface.DepthItemType{Price: 1.93}) == nil {
		t.Errorf("GetDepthBidQtyLocalMaxima returned an incorrect max bid price")
	}
	if bidLocalsMaxima.Get(&depth_interface.DepthItemType{Price: 1.949}) == nil {
		t.Errorf("GetDepthBidQtyLocalMaxima returned an incorrect max bid price")
	}
	if askLocalMaxima.Get(&depth_interface.DepthItemType{Price: 1.951}) == nil {
		t.Errorf("GetDepthAskQtyLocalMaxima returned an incorrect max ask price")
	}
	if askLocalMaxima.Get(&depth_interface.DepthItemType{Price: 1.953}) == nil {
		t.Errorf("GetDepthAskQtyLocalMaxima returned an incorrect max ask price")
	}
	if askLocalMaxima.Get(&depth_interface.DepthItemType{Price: 1.957}) == nil {
		t.Errorf("GetDepthAskQtyLocalMaxima returned an incorrect max ask price")
	}
	bidLocalsMaxima.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		if (item.Price != 1.93) && (item.Price != 1.949) {
			t.Errorf("GetDepthBidQtyLocalMaxima returned an incorrect max bid price")
		}
		return true
	})
	askLocalMaxima.Ascend(func(i btree.Item) bool {
		item := i.(*depth_interface.DepthItemType)
		if item.Price != 1.951 && item.Price != 1.953 && item.Price != 1.957 {
			t.Errorf("GetDepthAskQtyLocalMaxima returned an incorrect max ask price")
		}
		return true
	})

	// Add additional assertions if needed

}

// Add more test functions for other functions in the file if needed
