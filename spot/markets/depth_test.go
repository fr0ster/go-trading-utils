package markets_test

import (
	"math/rand"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/google/btree"
)

var testDepthTree = btree.New(2)

func getRandomPriceTree(tree *btree.BTree) *markets.DepthItem {
	items := make([]*markets.DepthItem, 0, tree.Len())
	tree.Ascend(func(i btree.Item) bool {
		items = append(items, i.(*markets.DepthItem))
		return true
	})

	return items[rand.Intn(len(items))]
}

func getTwoRandomPricesTree(tree *btree.BTree) (markets.Price, *markets.DepthItem, markets.Price, *markets.DepthItem) {
	items := make([]markets.DepthItem, 0, tree.Len())
	tree.Ascend(func(i btree.Item) bool {
		items = append(items, *i.(*markets.DepthItem))
		return true
	})

	index1, index2 := rand.Intn(len(items)), rand.Intn(len(items))
	for index1 == index2 { // ensure we have two different indices
		index2 = rand.Intn(len(items))
	}

	price1, price2 := items[index1].Price, items[index2].Price
	if price1 > price2 { // ensure the first price is less than the second
		price1, price2 = price2, price1
	}

	return price1, &items[index1], price2, &items[index2]
}

func TestInitDepthTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	err := markets.InitDepths(client, "BTCUSDT")
	testDepthTree = markets.GetDepths()
	if err != nil {
		t.Errorf("Failed to initialize depth tree: %v", err)
	}

	// Add more test cases here
}

func TestGetDepthTree(t *testing.T) {
	if testDepthTree == nil || testDepthTree.Len() == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		markets.InitDepths(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = markets.GetDepths()
	}
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	if testDepthTree == nil || testDepthTree.Len() == 0 {
		t.Errorf("GetDepthTree returned an empty map")
	}

	// Add additional assertions if needed
}

func TestSearchDepthTree(t *testing.T) {
	if testDepthTree == nil || testDepthTree.Len() == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		markets.InitDepths(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = markets.GetDepths()
	}

	// Declare and assign a value to the variable "price"
	price := markets.Price(10.0)

	// Call the function being tested
	filteredTree := markets.SearchDepths(price)

	// Add additional assertions to check the correctness of the returned ticker
	// For example, check if the ticker's symbol matches the expected value
	filteredTree.Ascend(func(i btree.Item) bool {
		price := i.(*markets.DepthItem)
		if price.Price != price.Price {
			t.Errorf("SearchDepthTreeByPrices returned a tree with incorrect prices")
		}
		return true
	})

	// Add additional assertions if needed
}

func TestSearchDepthTreeByPrices(t *testing.T) {
	if testDepthTree == nil || testDepthTree.Len() == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		markets.InitDepths(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = markets.GetDepths()
	}

	// Call the function being tested
	priceMin, _, priceMax, _ := getTwoRandomPricesTree(testDepthTree)
	markets.SetDepths(testDepthTree)
	filteredTree := markets.GetDepthsByPrices(priceMin, priceMax)

	// Add assertions to check the correctness of the filtered map
	filteredTree.Ascend(func(i btree.Item) bool {
		price := i.(*markets.DepthItem) // Modify the type assertion to use a pointer receiver
		if price.Price < priceMin || price.Price > priceMax {
			t.Errorf("SearchDepthTreeByPrices returned a tree with incorrect prices")
		}
		return true
	})

	// Add additional assertions if needed
}

func TestGetDepthMaxBidMinAsk(t *testing.T) {
	dataTree := btree.New(2)
	// Price: 1.947 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 172.1
	// Price: 1.948 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 187.4
	// Price: 1.949 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 236.1
	// Price: 1.95 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 189.8
	// Price: 1.951 AskLastUpdateID: 2369068 AskQuantity: 217.9 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.952 AskLastUpdateID: 2369068 AskQuantity: 179.4 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.953 AskLastUpdateID: 2369068 AskQuantity: 140.9 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.954 AskLastUpdateID: 2369068 AskQuantity: 148.5 BidLastUpdateID: 0 BidQuantity: 0
	records := []markets.DepthItem{
		{Price: 1.947, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 172.1},
		{Price: 1.948, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 187.4},
		{Price: 1.949, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 236.1},
		{Price: 1.95, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 189.8},
		{Price: 1.951, AskLastUpdateID: 2369068, AskQuantity: 217.9, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.952, AskLastUpdateID: 2369068, AskQuantity: 179.4, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.953, AskLastUpdateID: 2369068, AskQuantity: 140.9, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.954, AskLastUpdateID: 2369068, AskQuantity: 148.5, BidLastUpdateID: 0, BidQuantity: 0},
	}
	for _, record := range records {
		dataTree.ReplaceOrInsert(&record)
	}
	markets.SetDepths(dataTree)
	// Call the function being tested
	maxBid, minAsk := markets.GetDepthMaxBidMinAsk()
	if maxBid.Price != 1.95 {
		t.Errorf("GetDepthMaxBid returned an incorrect max bid price")
	}
	if minAsk.Price != 1.951 {
		t.Errorf("GetDepthMinAsk returned an incorrect min ask price")
	}
}

func TestGetDepthMaxBidQtyMaxAskQty(t *testing.T) {
	dataTree := btree.New(2)
	// Price: 1.947 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 172.1
	// Price: 1.948 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 187.4
	// Price: 1.949 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 236.1
	// Price: 1.95 AskLastUpdateID: 0 AskQuantity: 0 BidLastUpdateID: 2369068 BidQuantity: 189.8
	// Price: 1.951 AskLastUpdateID: 2369068 AskQuantity: 217.9 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.952 AskLastUpdateID: 2369068 AskQuantity: 179.4 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.953 AskLastUpdateID: 2369068 AskQuantity: 140.9 BidLastUpdateID: 0 BidQuantity: 0
	// Price: 1.954 AskLastUpdateID: 2369068 AskQuantity: 148.5 BidLastUpdateID: 0 BidQuantity: 0
	records := []markets.DepthItem{
		{Price: 1.947, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 172.1},
		{Price: 1.948, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 187.4},
		{Price: 1.949, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 236.1},
		{Price: 1.95, AskLastUpdateID: 0, AskQuantity: 0, BidLastUpdateID: 2369068, BidQuantity: 189.8},
		{Price: 1.951, AskLastUpdateID: 2369068, AskQuantity: 217.9, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.952, AskLastUpdateID: 2369068, AskQuantity: 179.4, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.953, AskLastUpdateID: 2369068, AskQuantity: 140.9, BidLastUpdateID: 0, BidQuantity: 0},
		{Price: 1.954, AskLastUpdateID: 2369068, AskQuantity: 148.5, BidLastUpdateID: 0, BidQuantity: 0},
	}
	for _, record := range records {
		dataTree.ReplaceOrInsert(&record)
	}
	markets.SetDepths(dataTree)
	// Call the function being tested
	maxBid, minAsk := markets.GetDepthMaxBidQtyMaxAskQty()
	if maxBid.Price != 1.949 {
		t.Errorf("GetDepthMaxBid returned an incorrect max bid price")
	}
	if minAsk.Price != 1.951 {
		t.Errorf("GetDepthMinAsk returned an incorrect min ask price")
	}
}

// Add more test functions for other functions in the file if needed
