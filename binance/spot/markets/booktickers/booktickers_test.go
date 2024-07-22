package bookticker_test

import (
	"errors"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_booktickers "github.com/fr0ster/go-trading-utils/binance/spot/markets/booktickers"
	bookticker_interface "github.com/fr0ster/go-trading-utils/interfaces/booktickers"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/booktickers/items"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func initBookTicker() *booktickers_types.BookTickers {
	bookTicker := booktickers_types.New(quit, 3, nil, nil)
	bookTicker.Set(bookticker_types.New("BTCUSDT", 10000, 1, 10001, 1))
	bookTicker.Set(bookticker_types.New("ETHUSDT", 1000, 1, 1001, 1))
	bookTicker.Set(bookticker_types.New("BNBUSDT", 100, 1, 101, 1))
	bookTicker.Set(bookticker_types.New("SUSHIUSDT", 10000, 1, 10001, 1))
	bookTicker.Set(bookticker_types.New("LINKUSDT", 1000, 1, 1001, 1))
	bookTicker.Set(bookticker_types.New("DOTUSDT", 100, 1, 101, 1))
	bookTicker.Set(bookticker_types.New("ADAUSDT", 10000, 1, 10001, 1))
	bookTicker.Set(bookticker_types.New("XRPUSDT", 1000, 1, 1001, 1))
	return bookTicker
}

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = true
	spot := binance.NewClient(api_key, secret_key)

	// Call the function under test
	bookTickers := booktickers_types.New(
		quit,
		3,
		nil,
		spot_booktickers.GetInitCreator(spot),
		"BTCUSDT")

	// TODO: Add more assertions to validate the behavior of the function
	btc_bt := bookTickers.Get("BTCUSDT")
	assert.NotNil(t, btc_bt)
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
	bookTicker.Set(bookticker_types.New("BTCUSDT", 10000, 1, 10001, 1))
	item := bookTicker.Get("BTCUSDT")
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.GetBidPrice() != 10000 && item.GetBidQuantity() != 1 && item.GetAskPrice() != 10001 && item.GetAskQuantity() != 1 {
		t.Errorf("SetItem did not update the item correctly")
	}

	bt := bookTicker.Get("BTCUSDT")
	bt.SetBidPrice(99999)
	bookTicker.Set(bt)
	bookTicker.Get("BTCUSDT")
	if item == nil {
		t.Errorf("SetItem did not update the item")
	} else if item.GetBidPrice() != 99999 {
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
