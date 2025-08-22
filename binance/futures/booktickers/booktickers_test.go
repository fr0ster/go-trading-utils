package booktickers_test

import (
	"testing"

	futures_booktickers "github.com/fr0ster/go-trading-utils/binance/futures/booktickers"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	booktickers_types "github.com/fr0ster/go-trading-utils/types/booktickers"

	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestInitPricesTree(t *testing.T) {
	t.Parallel()
	client := testutil.FuturesClient(t)

	// Call the function under test
	bookTicker := booktickers_types.New(quit, 3, nil, futures_booktickers.InitCreator(client), "BTCUSDT")

	// TODO: Add more assertions to validate the behavior of the function
	btc_bt := bookTicker.Get("BTCUSDT")
	assert.NotNil(t, btc_bt)
}
