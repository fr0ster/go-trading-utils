package trade_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/binance/spot/markets/trade"
	trade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestHistoricalTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	trades := trade_types.NewTrades()
	trade.HistoricalTradesInit(trades, binance.NewClient(api_key, secret_key), "BTCUSDT", 10)
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht, err := trade_types.Binance2Trades(item)
				assert.Nil(t, err)
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
	binance.UseTestnet = false
	trades := trade_types.NewTrades()
	trade.RecentTradesInit(trades, binance.NewClient(api_key, secret_key), "BTCUSDT", 10)
	test := func(i trade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
