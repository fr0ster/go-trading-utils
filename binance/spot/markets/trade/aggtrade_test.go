package trade_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/markets/trade"
	aggtrade_interface "github.com/fr0ster/go-trading-utils/interfaces/trades"
	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	aggTrade := trade.NewAggTrades()
	test := func(i aggtrade_interface.Trades) {

	}
	assert.NotPanics(t, func() {
		test(aggTrade)
	})
}

// func TestNewAggTrades(t *testing.T) {
// 	// Initialize the AggTrades
// 	aggTrade := trade.NewAggTrades()

// 	// Create some sample aggtrade items
// 	aggTradeItem1 := aggtrade_interface.AggTradeItem{
// 		AggTradeID:       1,
// 		Price:            "1.0",
// 		Quantity:         "1.0",
// 		FirstTradeID:     1,
// 		LastTradeID:      1,
// 		Timestamp:        1,
// 		IsBuyerMaker:     true,
// 		IsBestPriceMatch: true,
// 	}
// 	aggTradeItem2 := aggtrade_interface.AggTradeItem{
// 		AggTradeID:       2,
// 		Price:            "2.0",
// 		Quantity:         "2.0",
// 		FirstTradeID:     2,
// 		LastTradeID:      2,
// 		Timestamp:        2,
// 		IsBuyerMaker:     false,
// 		IsBestPriceMatch: false,
// 	}

// 	// Set the aggtrade items in the tree
// 	aggTrade.Set(aggTradeItem1)
// 	aggTrade.Set(aggTradeItem2)

// 	// Get the aggtrade item by aggtradeID
// 	result, err := aggTrade.Get(1)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if result != aggTradeItem1 {
// 		t.Error(errors.New("AggTradeItem1 not found"))
// 	}
// }
