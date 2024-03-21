package trade_test

import (
	"os"
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/markets/trade"
	aggtrade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	"github.com/stretchr/testify/assert"
)

func TestListTradeInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade.NewTradesV3()
	trade.ListTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestListMarginTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade.NewTradesV3()
	trade.ListMarginTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestHistoricalTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade.NewTrades()
	trade.HistoricalTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestRecentTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade.NewTrades()
	trade.RecentTradesInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestAggTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	UseTestnet := false
	trades := trade.NewAggTrades()
	trade.AggTradeInit(trades, api_key, secret_key, "BTCUSDT", 10, UseTestnet)
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
