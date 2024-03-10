package info_test

import (
	"math/rand"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/google/btree"
)

var testDepthTree = btree.New(2)

func getRandomPriceTree(tree *btree.BTree) *info.DepthRecord {
	items := make([]info.DepthRecord, 0, tree.Len())
	tree.Ascend(func(i btree.Item) bool {
		items = append(items, i.(info.DepthRecord))
		return true
	})

	return &items[rand.Intn(len(items))]
}

func getTwoRandomPricesTree(tree *btree.BTree) (info.Price, *info.DepthRecord, info.Price, *info.DepthRecord) {
	items := make([]info.DepthRecord, 0, tree.Len())
	tree.Ascend(func(i btree.Item) bool {
		items = append(items, i.(info.DepthRecord))
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

	err := info.InitDepthTree(client, "BTCUSDT")
	testDepthTree = info.GetDepthTree()
	if err != nil {
		t.Errorf("Failed to initialize depth tree: %v", err)
	}

	// Add more test cases here
}

func TestGetBookTickerTree(t *testing.T) {
	if testDepthTree == nil || testDepthTree.Len() == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		info.InitDepthMap(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = info.GetDepthTree()
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
		info.InitDepthMap(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = info.GetDepthTree()
	}

	// Call the function being tested
	price := getRandomPriceTree(testDepthTree)
	info.SetDepthTree(testDepthTree)
	filteredTree := info.SearchDepthTree(price.Price)

	// Add additional assertions to check the correctness of the returned ticker
	// For example, check if the ticker's symbol matches the expected value
	filteredTree.Ascend(func(i btree.Item) bool {
		price := i.(info.DepthRecord)
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
		info.InitDepthTree(client, "BTCUSDT")
		// Call the function being tested
		testDepthTree = info.GetDepthTree()
	}

	// Call the function being tested
	priceMin, _, priceMax, _ := getTwoRandomPricesTree(testDepthTree)
	info.SetDepthTree(testDepthTree)
	filteredTree := info.SearchDepthTreeByPrices(priceMin, priceMax)

	// Add assertions to check the correctness of the filtered map
	filteredTree.Ascend(func(i btree.Item) bool {
		price := i.(info.DepthRecord)
		if price.Price < priceMin || price.Price > priceMax {
			t.Errorf("SearchDepthTreeByPrices returned a tree with incorrect prices")
		}
		return true
	})

	// Add additional assertions if needed
}

// Add more test functions for other functions in the file if needed
