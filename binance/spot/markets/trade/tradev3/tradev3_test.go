package trade_test

import (
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	spot_trade "github.com/fr0ster/go-trading-utils/binance/spot/markets/trade/tradev3"
	trade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	trade_types "github.com/fr0ster/go-trading-utils/types/trade/tradeV3"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func TestListTradeInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	trades := trade_types.New(
		"BTCUSDT",
		nil,
		spot_trade.GetListTradesInitCreator(binance.NewClient(api_key, secret_key), 10))
	test := func(i trade_interface.Trades) {
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

func TestListMarginTradesInterface(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	binance.UseTestnet = false
	trades := trade_types.New(
		"BTCUSDT",
		nil,
		spot_trade.GetListMarginTradesInitCreator(binance.NewClient(api_key, secret_key), 10))
	test := func(i trade_interface.Trades) {
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
