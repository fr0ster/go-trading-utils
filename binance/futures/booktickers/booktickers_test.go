package booktickers_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"

	futures_booktickers "github.com/fr0ster/go-trading-utils/binance/futures/booktickers"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"

	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestInitPricesTree(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = true
	futures := futures.NewClient(api_key, secret_key)

	// Call the function under test
	bookTicker := booktickers_types.New(quit, 3, nil, futures_booktickers.InitCreator(futures), "BTCUSDT")

	// TODO: Add more assertions to validate the behavior of the function
	btc_bt := bookTicker.Get("BTCUSDT")
	assert.NotNil(t, btc_bt)
}
