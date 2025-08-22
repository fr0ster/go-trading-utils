package bookticker_test

import (
	"testing"

	spot_booktickers "github.com/fr0ster/go-trading-utils/binance/spot/booktickers"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestInitPricesTree(t *testing.T) {
	t.Parallel()
	client := testutil.SpotClient(t)

	// Call the function under test
	bookTickers := booktickers_types.New(
		quit,
		3,
		nil,
		spot_booktickers.InitCreator(client),
		"BTCUSDT")

	// TODO: Add more assertions to validate the behavior of the function
	btc_bt := bookTickers.Get("BTCUSDT")
	assert.NotNil(t, btc_bt)
}
