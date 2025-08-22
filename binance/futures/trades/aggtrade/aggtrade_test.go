package aggtrade_test

import (
	"testing"

	futures_trade "github.com/fr0ster/go-trading-utils/binance/futures/trades/aggtrade"
	"github.com/fr0ster/go-trading-utils/internal/testutil"
	trade_types "github.com/fr0ster/go-trading-utils/types/trades/aggtrade"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

var (
	quit = make(chan struct{})
)

func TestAggTrades(t *testing.T) {
	t.Parallel()
	trades := trade_types.New(
		quit,
		"BTCUSDT",
		futures_trade.TradeStreamCreator(nil, nil),
		futures_trade.InitCreator(testutil.FuturesClient(t), 10))
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
