package trade_test

import (
	"testing"

	futures_trade "github.com/fr0ster/go-trading-utils/binance/futures/trades/trade"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/trade"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestHistoricalTrades(t *testing.T) {
	t.Parallel()
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		futures_trade.HistoricalTradesInitCreator(testutil.FuturesClient(t), 10))
	test := func(i *trade_types.Trades) {
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

func TestRecentTrades(t *testing.T) {
	t.Parallel()
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		futures_trade.RecentTradesInitCreator(testutil.FuturesClient(t), 10))
	test := func(i *trade_types.Trades) {
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
