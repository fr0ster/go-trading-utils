package pairs_test

import (
	"math"
	"testing"

	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func getTestData() *btree.BTree {
	res := btree.New(2)
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:              pairs_types.SpotAccountType,
		StrategyType:             pairs_types.HoldingStrategyType,
		StageType:                pairs_types.InputIntoPositionStage,
		Pair:                     "BTCUSDT",
		MarginType:               pairs_types.CrossMarginType,
		Leverage:                 20,
		LimitInputIntoPosition:   0.5,
		LimitOutputOfPosition:    0.8,
		LimitOnPosition:          0.9,
		LimitOnTransaction:       0.1,
		UnRealizedProfitLowBound: 0.1,
		UnRealizedProfitUpBound:  0.9,
		DeltaPrice:               0.01,
		DeltaQuantity:            0.1,
		IsArithmetic:             true,
		BuyQuantity:              0.3,
		BuyValue:                 300.0,
		SellQuantity:             0.2,
		SellValue:                200.0,
		CallbackRate:             0.1, // CallbackRate 0.1%
	})
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:              pairs_types.USDTFutureType,
		StrategyType:             pairs_types.ScalpingStrategyType,
		StageType:                pairs_types.WorkInPositionStage,
		Pair:                     "BTCUSDT",
		MarginType:               pairs_types.CrossMarginType,
		Leverage:                 20,
		LimitInputIntoPosition:   0.5,
		LimitOutputOfPosition:    0.8,
		LimitOnPosition:          0.9,
		LimitOnTransaction:       0.1,
		UnRealizedProfitLowBound: 0.1,
		UnRealizedProfitUpBound:  0.9,
		DeltaPrice:               0.01,
		DeltaQuantity:            0.1,
		IsArithmetic:             true,
		BuyQuantity:              0.3,
		BuyValue:                 300.0,
		SellQuantity:             0.2,
		SellValue:                200.0,
		CallbackRate:             0.1, // CallbackRate 0.1%
	})
	return res
}

func assertPair(
	t *testing.T,
	pair *pairs_types.Pairs,
	accountType pairs_types.AccountType,
	strategyType pairs_types.StrategyType,
	stageType pairs_types.StageType) {

	// Test GetAccountType
	assert.Equal(t, accountType, pair.GetAccountType())

	// Test GetStrategy
	assert.Equal(t, strategyType, pair.GetStrategy())

	// Test GetStage
	assert.Equal(t, stageType, pair.GetStage())

	// Test SetStage
	pair.SetStage(pairs_types.WorkInPositionStage)
	assert.Equal(t, pairs_types.WorkInPositionStage, pair.GetStage())

	// Test GetPair
	assert.Equal(t, "BTCUSDT", pair.GetPair())

	// Test GetMarginType
	assert.Equal(t, pairs_types.CrossMarginType, pair.GetMarginType())

	// Test GetLeverage
	assert.Equal(t, 20, pair.GetLeverage())

	// Test GetLimitInputIntoPosition
	assert.Equal(t, 0.5, pair.GetLimitInputIntoPosition())

	// Test GetLimitOutputOfPosition
	assert.Equal(t, 0.8, pair.GetLimitOutputOfPosition())

	// Test GetLimitOnPosition
	assert.Equal(t, 0.9, pair.GetLimitOnPosition())

	// Test GetLimitOnTransaction
	assert.Equal(t, 0.1, pair.GetLimitOnTransaction())

	// Test GetDeltaPrice
	assert.Equal(t, 0.01, pair.GetDeltaPrice())

	// Test GetDeltaQuantity
	assert.Equal(t, 0.1, pair.GetDeltaQuantity())

	// Test GetIsArithmetic
	assert.True(t, pair.GetIsArithmetic())

	// Test GetBuyQuantity
	assert.Equal(t, 0.3, pair.GetBuyQuantity())

	// Test GetSellQuantity
	assert.Equal(t, 0.2, pair.GetSellQuantity())

	// Test GetBuyValue
	assert.Equal(t, 300.0, pair.GetBuyValue())

	// Test GetSellValue
	assert.Equal(t, 200.0, pair.GetSellValue())

	// Test GetCallbackRate
	assert.Equal(t, 0.1, pair.GetCallbackRate())

	// Test GetMiddlePrice
	assert.Equal(t, 1000.0, math.Round(pair.GetMiddlePrice()))

	// Test GetProfit
	profit := pair.GetProfit(60000.0)
	assert.Equal(t, 5900.0, math.Round(profit))

	// Test CheckingPair
	result := pair.CheckingPair()
	assert.True(t, result)
}

func TestPairs(t *testing.T) {
	pairs := getTestData()
	val := pairs.Get(&pairs_types.Pairs{
		AccountType:  pairs_types.SpotAccountType,
		StrategyType: pairs_types.HoldingStrategyType,
		StageType:    pairs_types.InputIntoPositionStage,
		Pair:         "BTCUSDT"})
	assert.NotNil(t, val)
	pair_1 := val.(*pairs_types.Pairs)
	assertPair(t, pair_1, pairs_types.SpotAccountType, pairs_types.HoldingStrategyType, pairs_types.InputIntoPositionStage)
	val = pairs.Get(&pairs_types.Pairs{
		AccountType:  pairs_types.USDTFutureType,
		StrategyType: pairs_types.ScalpingStrategyType,
		StageType:    pairs_types.WorkInPositionStage,
		Pair:         "BTCUSDT"})
	assert.NotNil(t, val)
	pair_2 := val.(*pairs_types.Pairs)
	assertPair(t, pair_2, pairs_types.USDTFutureType, pairs_types.ScalpingStrategyType, pairs_types.WorkInPositionStage)
}
