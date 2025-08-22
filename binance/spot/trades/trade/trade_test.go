package trade_test

import (
	"testing"

	spot_trade "github.com/fr0ster/go-trading-utils/binance/spot/trades/trade"
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
		spot_trade.HistoricalTradesInitCreator(testutil.SpotClient(t), 10))
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
		spot_trade.RecentTradesInitCreator(testutil.SpotClient(t), 10))
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
