package trade_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/binance/futures/markets/trade"
	trade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestAggTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	trades := trade_types.NewAggTrades()
	trade.AggTradeInit(trades, futures.NewClient(api_key, secret_key), "BTCUSDT", 10)
	test := func(i trade_interface.Trades) {
		i.Lock()
		defer i.Unlock()
		i.Ascend(func(item btree.Item) bool {
			if item != nil {
				ht, err := trade_types.Binance2AggTrades(item)
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
