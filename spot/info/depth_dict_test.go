package info

import (
	"math/rand"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
)

var BookTickerMap = make(map[Price]BookTicker)

func getRandomPrice(dict map[Price]BookTicker) (Price, BookTicker) {
	keys := make([]Price, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	randomKey := keys[rand.Intn(len(keys))]
	return randomKey, dict[randomKey]
}

func getTwoRandomPrices(dict map[Price]BookTicker) (Price, BookTicker, Price, BookTicker) {
	keys := make([]Price, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}

	index1, index2 := rand.Intn(len(keys)), rand.Intn(len(keys))
	for index1 == index2 { // ensure we have two different indices
		index2 = rand.Intn(len(keys))
	}

	key1, key2 := keys[index1], keys[index2]
	if key1 > key2 { // ensure the first price is less than the second
		key1, key2 = key2, key1
	}

	return key1, dict[key1], key2, dict[key2]
}

func TestInitDepthDictMap(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the function being tested
	err := InitDepthDictMap(client, "BTCUSDT")
	bookTickerMap = GetBookTickerMap()

	// Check if there was an error
	if err != nil {
		t.Errorf("InitDepthDictMap returned an error: %v", err)
	}

	// Add additional assertions if needed
}

func TestGetBookTickerMap(t *testing.T) {
	if len(BookTickerMap) == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		InitDepthDictMap(client, "BTCUSDT")
		// Call the function being tested
		bookTickerMap = GetBookTickerMap()
	}
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	if len(bookTickerMap) == 0 {
		t.Errorf("GetBookTickerMap returned an empty map")
	}

	// Add additional assertions if needed
}

func TestSearchBookTickerMap(t *testing.T) {
	if len(BookTickerMap) == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		InitDepthDictMap(client, "BTCUSDT")
		// Call the function being tested
		bookTickerMap = GetBookTickerMap()
	}

	// Call the function being tested
	price, _ := getRandomPrice(bookTickerMap)
	SetBookTickerMap(bookTickerMap)
	ticker, found := SearchBookTickerMap(price)

	// Check if the ticker was found
	if !found {
		t.Errorf("SearchBookTickerMap did not find the ticker")
	}

	// Add additional assertions to check the correctness of the returned ticker
	// For example, check if the ticker's symbol matches the expected value
	if ticker.Price != price {
		t.Errorf("SearchBookTickerMap returned ticker with incorrect price: %v", ticker.Price)
	}

	// Add additional assertions if needed
}

func TestSearchBookTickerMapByPrices(t *testing.T) {
	if len(BookTickerMap) == 0 {
		api_key := os.Getenv("API_KEY")
		secret_key := os.Getenv("SECRET_KEY")
		binance.UseTestnet = true
		client := binance.NewClient(api_key, secret_key)
		InitDepthDictMap(client, "BTCUSDT")
		// Call the function being tested
		bookTickerMap = GetBookTickerMap()
	}

	// Call the function being tested
	priceMin, _, priceMax, _ := getTwoRandomPrices(bookTickerMap)
	SetBookTickerMap(bookTickerMap)
	filteredMap := SearchBookTickerMapByPrices(priceMin, priceMax)

	// Add assertions to check the correctness of the filtered map
	for key := range filteredMap {
		if key < priceMin || key > priceMax {
			t.Errorf("SearchBookTickerMapByPrices returned a map with incorrect prices")
		}
	}

	// Add additional assertions if needed
}
