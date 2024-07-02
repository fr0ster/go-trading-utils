package pairs_test

import (
	"testing"

	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func getTestData() *btree.BTree {
	res := btree.New(2)
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:        pairs_types.SpotAccountType,
		StrategyType:       pairs_types.HoldingStrategyType,
		StageType:          pairs_types.InputIntoPositionStage,
		Pair:               "BTCUSDT",
		MarginType:         pairs_types.CrossMarginType,
		Leverage:           20,
		LimitOnPosition:    1000,
		LimitOnTransaction: 10,   // LimitOnTransaction 10%
		UpBound:            10,   // UpBoundPercent 10%
		LowBound:           10,   // LowBoundPercent 10%
		DeltaPrice:         1.0,  // DeltaPrice 1%
		DeltaQuantity:      10.0, // DeltaQuantity 10%
		Progression:        "GEOMETRIC",
		Value:              200.0,
		CallbackRate:       0.1, // CallbackRate 0.1%
	})
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:        pairs_types.USDTFutureType,
		StrategyType:       pairs_types.ScalpingStrategyType,
		StageType:          pairs_types.WorkInPositionStage,
		Pair:               "BTCUSDT",
		MarginType:         pairs_types.CrossMarginType,
		Leverage:           20,
		LimitOnPosition:    1000,
		LimitOnTransaction: 10,   // LimitOnTransaction 10%
		UpBound:            10,   // UpBoundPercent 10%
		LowBound:           10,   // LowBoundPercent 10%
		DeltaPrice:         1.0,  // DeltaPrice 1%
		DeltaQuantity:      10.0, // DeltaQuantity 10%
		Progression:        "GEOMETRIC",
		Value:              200.0,
		CallbackRate:       0.1, // CallbackRate 0.1%
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

	// Test GetLimitOnPosition
	assert.Equal(t, 1000.0, pair.GetLimitOnPosition())

	// Test GetLimitOnTransaction
	assert.Equal(t, 0.1, pair.GetLimitOnTransaction())

	// Test GetUpBoundPercent
	assert.Equal(t, 0.1, pair.GetUpBound())

	// Test GetLowBoundPercent
	assert.Equal(t, 0.1, pair.GetLowBound())

	// Test GetDeltaPrice
	assert.Equal(t, 0.01, pair.GetDeltaPrice())

	// Test GetDeltaQuantity
	assert.Equal(t, 0.1, pair.GetDeltaQuantity())

	// Test GetProgression
	assert.Equal(t, pairs_types.GeometricProgression, pair.GetProgression())

	// Test GetValue
	assert.Equal(t, 200.0, pair.GetValue())

	// Test GetCallbackRate
	assert.Equal(t, 0.1, pair.GetCallbackRate())
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
