package trade_test

import (
	"testing"

	spot_trade "github.com/fr0ster/go-trading-utils/binance/spot/trades/tradev3"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/tradeV3"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestListTrade(t *testing.T) {
	t.Parallel()
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		spot_trade.ListTradesInitCreator(testutil.SpotClient(t), 10))
	test := func(i *trade_types.TradesV3) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(*trade_types.TradeV3)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}

func TestListMarginTrades(t *testing.T) {
	t.Parallel()
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		nil,
		spot_trade.ListMarginTradesInitCreator(testutil.SpotClient(t), 10))
	test := func(i *trade_types.TradesV3) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht := item.(*trade_types.TradeV3)
				assert.NotNil(t, ht)
			}
			return true
		})
	}
	assert.NotPanics(t, func() {
		test(trades)
	})
}
