package bookticker_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_booktickers "github.com/fr0ster/go-trading-utils/binance/spot/booktickers"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

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
		spot_booktickers.InitCreator(spot),
		"BTCUSDT")

	// TODO: Add more assertions to validate the behavior of the function
	btc_bt := bookTickers.Get("BTCUSDT")
	assert.NotNil(t, btc_bt)
}
