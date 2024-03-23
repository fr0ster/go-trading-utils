package bookticker_test

import (
	"errors"
	"os"
	"testing"

	futuresBookTicker "github.com/fr0ster/go-trading-utils/binance/spot/markets/bookticker"
	bookticker_interface "github.com/fr0ster/go-trading-utils/interfaces/bookticker"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
)

func initBookTicker() *bookticker_types.BookTickerBTree {
	bookTicker := bookticker_types.New(3)
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "BTCUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "ETHUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "BNBUSDT", BidPrice: 100, BidQuantity: 1, AskPrice: 101, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "SUSHIUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "LINKUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "DOTUSDT", BidPrice: 100, BidQuantity: 1, AskPrice: 101, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "ADAUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "XRPUSDT", BidPrice: 1000, BidQuantity: 1, AskPrice: 1001, AskQuantity: 1})
	return bookTicker
}

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	// binance.UseTestnet = true

	// Call the function under test
	bookTicker := bookticker_types.New(3)
	err := futuresBookTicker.Init(bookTicker, api_key, secret_key, "BTCUSDT", false)

	// Check if there was an error
	if err != nil {
		t.Errorf("InitPricesTree returned an error: %v", err)
	}

	// TODO: Add more assertions to validate the behavior of the function
}

func TestBookTickerGet(t *testing.T) {
	// Add assertions to check the correctness of the returned item
	// For example, check if the item is not nil
	bookTicker := initBookTicker()
	item := bookTicker.Get("BTCUSDT")
	if item == nil {
		t.Errorf("GetItem returned an empty item")
	}
}

func TestSetBookTickerItem(t *testing.T) {
	// Add assertions to check the correctness of the updated item
	// For example, check if the item was updated correctly
	bookTicker := initBookTicker()
	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "BTCUSDT", BidPrice: 10000, BidQuantity: 1, AskPrice: 10001, AskQuantity: 1})
	item, err := bookticker_types.Binance2BookTicker(bookTicker.Get("BTCUSDT"))
	if err != nil {
		t.Errorf("SetItem returned an error: %v", err)
	}
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.BidPrice != 10000 && item.BidQuantity != 1 && item.AskPrice != 10001 && item.AskQuantity != 1 {
		t.Errorf("SetItem did not update the item correctly")
	}

	bookTicker.Set(&bookticker_types.BookTickerItem{Symbol: "BTCUSDT", BidPrice: 99999})
	item, err = bookticker_types.Binance2BookTicker(bookTicker.Get("BTCUSDT"))
	if err != nil {
		t.Errorf("SetItem returned an error: %v", err)
	}
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.BidPrice != 99999 {
		t.Errorf("SetItem did not update the item correctly")
	}
}

func TestInterface(t *testing.T) {
	btt := initBookTicker()
	err := func(val bookticker_interface.BookTicker) error {
		item := val.Get("BTCUSDT")
		if item == nil {
			return errors.New("GetItem returned an empty item")
		}
		return nil
	}(btt)
	if err != nil {
		t.Error(err)
	}
}
