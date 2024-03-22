package trade_test

import (
	"os"
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/futures/markets/trade"
	trade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestHistoricalTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade_types.NewTrades()
	trade.HistoricalTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(trade_types.Trade)
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
	UseTestnet := false
	trades := trade_types.NewTrades()
	trade.RecentTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(trade_types.Trade)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
