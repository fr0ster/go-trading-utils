package aggtrade_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"

	spot_trade "github.com/fr0ster/go-trading-utils/binance/spot/trades/aggtrade"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/aggtrade"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestAggTrades(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		spot_trade.TradeStreamCreator(nil, nil),
		spot_trade.InitCreator(binance.NewClient(api_key, secret_key), 10))
	test := func(i *trade_types.AggTrades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(*trade_types.AggTrade)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
