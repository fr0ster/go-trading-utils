package markets_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/markets"
	"github.com/google/btree"
)

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the function under test
	err := markets.InitBookTicker(client, "BTCUSDT")

	// Check if there was an error
	if err != nil {
		t.Errorf("InitPricesTree returned an error: %v", err)
	}

	// TODO: Add more assertions to validate the behavior of the function
}

func TestGetBookTickerTree(t *testing.T) {
	// Call the function under test
	tree := markets.GetBookTickers()

	// TODO: Add assertions to validate the behavior of the function
	if tree == nil {
		t.Errorf("Expected non-nil tree, got nil")
	}
}

func TestSetBookTickerTree(t *testing.T) {
	// Create a new BTree
	tree := btree.New(2)
	tree.ReplaceOrInsert(markets.BookTickerItem{
		Symbol:      markets.SymbolType("BTCUSDT"),
		BidPrice:    markets.PriceType(10000),
		BidQuantity: markets.PriceType(1),
		AskPrice:    markets.PriceType(10001),
		AskQuantity: markets.PriceType(1),
	})
	tree.ReplaceOrInsert(markets.BookTickerItem{
		Symbol:      markets.SymbolType("ETHUSDT"),
		BidPrice:    markets.PriceType(200),
		BidQuantity: markets.PriceType(2),
		AskPrice:    markets.PriceType(201),
		AskQuantity: markets.PriceType(2),
	})

	// Call the function under test
	markets.SetBookTickers(tree)

	// TODO: Add assertions to validate the behavior of the function
	if markets.GetBookTickers() != tree {
		t.Errorf("Expected tree: %v, got: %v", tree, markets.GetBookTickers())
	}
}

// TODO: Add more tests for the other functions
