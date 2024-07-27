package pairs_test

import (
	"testing"

	types "github.com/fr0ster/go-trading-utils/types"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func getTestData() *btree.BTree {
	res := btree.New(2)
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:        types.SpotAccountType,
		StrategyType:       types.HoldingStrategyType,
		StageType:          types.InputIntoPositionStage,
		Pair:               "BTCUSDT",
		MarginType:         types.CrossMarginType,
		Leverage:           20,
		LimitOnPosition:    1000,
		LimitOnTransaction: 10,   // LimitOnTransaction 10%
		UpAndLowBound:      10,   // UpBoundPercent 10%
		DeltaPrice:         1.0,  // DeltaPrice 1%
		DeltaQuantity:      10.0, // DeltaQuantity 10%
		Progression:        "GEOMETRIC",
		Value:              200.0,
		CallbackRate:       0.1, // CallbackRate 0.1%
		DepthsN:            50,  // DepthsN 50
	})
	res.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:        types.USDTFutureType,
		StrategyType:       types.ScalpingStrategyType,
		StageType:          types.WorkInPositionStage,
		Pair:               "BTCUSDT",
		MarginType:         types.CrossMarginType,
		Leverage:           20,
		LimitOnPosition:    1000,
		LimitOnTransaction: 10,   // LimitOnTransaction 10%
		UpAndLowBound:      10,   // UpBoundPercent 10%
		DeltaPrice:         1.0,  // DeltaPrice 1%
		DeltaQuantity:      10.0, // DeltaQuantity 10%
		Progression:        "GEOMETRIC",
		Value:              200.0,
		CallbackRate:       0.1, // CallbackRate 0.1%
		DepthsN:            50,  // DepthsN 50
	})
	return res
}

func assertPair(
	t *testing.T,
	pair *pairs_types.Pairs,
	accountType types.AccountType,
	strategyType types.StrategyType,
	stageType types.StageType) {

	// Test GetAccountType
	assert.Equal(t, accountType, pair.GetAccountType())

	// Test GetStrategy
	assert.Equal(t, strategyType, pair.GetStrategy())

	// Test GetStage
	assert.Equal(t, stageType, pair.GetStage())

	// Test SetStage
	pair.SetStage(types.WorkInPositionStage)
	assert.Equal(t, types.WorkInPositionStage, pair.GetStage())

	// Test GetPair
	assert.Equal(t, "BTCUSDT", pair.GetPair())

	// Test GetMarginType
	assert.Equal(t, types.CrossMarginType, pair.GetMarginType())

	// Test GetLeverage
	assert.Equal(t, 20, pair.GetLeverage())

	// Test GetLimitOnPosition
	assert.Equal(t, items_types.ValueType(1000.0), pair.GetLimitOnPosition())

	// Test GetLimitOnTransaction
	assert.Equal(t, items_types.ValuePercentType(10), pair.GetLimitOnTransaction())

	// Test GetUpBoundPercent
	assert.Equal(t, items_types.PricePercentType(10), pair.GetUpAndLowBound())

	// Test GetDeltaPrice
	assert.Equal(t, items_types.PricePercentType(1), pair.GetDeltaPrice())

	// Test GetDeltaQuantity
	assert.Equal(t, items_types.QuantityPercentType(10.0), pair.GetDeltaQuantity())

	// Test GetProgression
	assert.Equal(t, types.GeometricProgression, pair.GetProgression())

	// Test GetValue
	assert.Equal(t, items_types.ValueType(200.0), pair.GetValue())

	// Test GetCallbackRate
	assert.Equal(t, items_types.PricePercentType(0.1), pair.GetCallbackRate())

	// Test GetDepthsN
	assert.Equal(t, depth_types.DepthAPILimit(50), pair.GetDepthsN())
}

func TestPairs(t *testing.T) {
	pairs := getTestData()
	val := pairs.Get(&pairs_types.Pairs{
		AccountType:  types.SpotAccountType,
		StrategyType: types.HoldingStrategyType,
		StageType:    types.InputIntoPositionStage,
		Pair:         "BTCUSDT"})
	assert.NotNil(t, val)
	pair_1 := val.(*pairs_types.Pairs)
	assertPair(t, pair_1, types.SpotAccountType, types.HoldingStrategyType, types.InputIntoPositionStage)
	val = pairs.Get(&pairs_types.Pairs{
		AccountType:  types.USDTFutureType,
		StrategyType: types.ScalpingStrategyType,
		StageType:    types.WorkInPositionStage,
		Pair:         "BTCUSDT"})
	assert.NotNil(t, val)
	pair_2 := val.(*pairs_types.Pairs)
	assertPair(t, pair_2, types.USDTFutureType, types.ScalpingStrategyType, types.WorkInPositionStage)
}
