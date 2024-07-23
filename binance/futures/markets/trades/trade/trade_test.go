package trade_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	futures_trade "github.com/fr0ster/go-trading-utils/binance/futures/markets/trades/trade"
	trade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/trade"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestHistoricalTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		futures_trade.HistoricalTradesInitCreator(futures.NewClient(api_key, secret_key), 10))
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(*trade_types.Trade)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestRecentTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		futures_trade.RecentTradesInitCreator(futures.NewClient(api_key, secret_key), 10))
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(*trade_types.Trade)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
