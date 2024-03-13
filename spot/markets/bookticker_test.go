package markets_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/google/btree"
)

func initBookTicker() *markets.BookTickerBTree {
	bookTicker := markets.BookTickerNew(3)
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "BTCUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "ETHUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "BNBUSDT", BidPrice: 100, BidQuantity: 1, AskPrice: 101, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "SUSHIUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "LINKUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "DOTUSDT", BidPrice: 100, BidQuantity: 1, AskPrice: 101, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "ADAUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.ReplaceOrInsert(&markets.BookTickerItemType{Symbol: "XRPUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	return bookTicker
}

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the function under test
	bookTicker := markets.BookTickerNew(3)
	err := bookTicker.Init(client, "BTCUSDT")

	// Check if there was an error
	if err != nil {
		t.Errorf("InitPricesTree returned an error: %v", err)
	}

	// TODO: Add more assertions to validate the behavior of the function
}

func TestBookTickerGetItem(t *testing.T) {
	// Add assertions to check the correctness of the returned item
	// For example, check if the item is not nil
	bookTicker := initBookTicker()
	item := bookTicker.GetItem("BTCUSDT")
	if item == nil {
		t.Errorf("GetItem returned an empty item")
	}
}

func TestSetBookTickerItem(t *testing.T) {
	// Add assertions to check the correctness of the updated item
	// For example, check if the item was updated correctly
	bookTicker := initBookTicker()
	bookTicker.SetItem(markets.BookTickerItemType{Symbol: "BTCUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	item := bookTicker.GetItem("BTCUSDT")
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.BidPrice != 10000 && item.BidQuantity != 1 && item.AskPrice != 10001 && item.AskQuantity != 1 {
		t.Errorf("SetItem did not update the item correctly")
	}

	bookTicker.SetItem(markets.BookTickerItemType{Symbol: "BTCUSDT", BidPrice: 99999})
	item = bookTicker.GetItem("BTCUSDT")
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.BidPrice != 99999 {
		t.Errorf("SetItem did not update the item correctly")
	}
}

func TestGetBookTickersBySymbol(t *testing.T) {
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	bookTicker := initBookTicker()
	newTree := bookTicker.GetBySymbol("BTCUSDT")
	if newTree == nil {
		t.Errorf("GetBySymbol returned an empty map")
	} else {
		newTree.Ascend(func(i btree.Item) bool {
			item := i.(*markets.BookTickerItemType)
			if item.Symbol != "BTCUSDT" {
				t.Errorf("GetBySymbol returned a map with incorrect symbols")
			}
			return true
		})
	}
}

func TestGetBookTickersByBidPrice(t *testing.T) {
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	bookTicker := initBookTicker()
	searchPrice := markets.PriceType(10000)
	newTree := bookTicker.GetByBidPrice("BTCUSDT", searchPrice)
	if newTree == nil {
		t.Errorf("GetByBidPrice returned an empty map")
	} else {
		newTree.Ascend(func(i btree.Item) bool {
			item := i.(*markets.BookTickerItemType)
			if item.BidPrice != searchPrice {
				t.Errorf("GetByBidPrice returned a map with incorrect bid prices")
			}
			return true
		})
	}
}

func TestGetBookTickerByAskPrice(t *testing.T) {
	// Add assertions to check the correctness of the returned map
	// For example, check if the map is not empty
	bookTicker := initBookTicker()
	searchPrice := markets.PriceType(10001)
	newTree := bookTicker.GetByAskPrice("BTCUSDT", searchPrice)
	if newTree == nil {
		t.Errorf("GetByBidPrice returned an empty map")
	} else {
		newTree.Ascend(func(i btree.Item) bool {
			item := i.(*markets.BookTickerItemType)
			if item.AskPrice != searchPrice {
				t.Errorf("GetByAskPrice returned a map with incorrect ask prices")
			}
			return true
		})
	}
}

// TODO: Add more tests for the other functions
