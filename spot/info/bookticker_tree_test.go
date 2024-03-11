package info_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/info"
	"github.com/google/btree"
)

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	// Call the function under test
	err := info.InitPricesTree(client, "BTCUSDT")

	// Check if there was an error
	if err != nil {
		t.Errorf("InitPricesTree returned an error: %v", err)
	}

	// TODO: Add more assertions to validate the behavior of the function
}

func TestGetBookTickerTree(t *testing.T) {
	// Call the function under test
	tree := info.GetBookTickerTree()

	// TODO: Add assertions to validate the behavior of the function
	if tree == nil {
		t.Errorf("Expected non-nil tree, got nil")
	}
}

func TestSetBookTickerTree(t *testing.T) {
	// Create a new BTree
	tree := btree.New(2)
	tree.ReplaceOrInsert(info.BookTickerItem{
		Symbol:      info.SymbolName("BTCUSDT"),
		BidPrice:    info.SymbolPrice(10000),
		BidQuantity: info.SymbolPrice(1),
		AskPrice:    info.SymbolPrice(10001),
		AskQuantity: info.SymbolPrice(1),
	})
	tree.ReplaceOrInsert(info.BookTickerItem{
		Symbol:      info.SymbolName("ETHUSDT"),
		BidPrice:    info.SymbolPrice(200),
		BidQuantity: info.SymbolPrice(2),
		AskPrice:    info.SymbolPrice(201),
		AskQuantity: info.SymbolPrice(2),
	})

	// Call the function under test
	info.SetBookTickerTree(tree)

	// TODO: Add assertions to validate the behavior of the function
	if info.GetBookTickerTree() != tree {
		t.Errorf("Expected tree: %v, got: %v", tree, info.GetBookTickerTree())
	}
}

// TODO: Add more tests for the other functions
