package depths_test

import (
	"testing"

	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	item_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	"github.com/stretchr/testify/assert"
)

const (
	degree = 3
)

func TestDepthsGetAndReplaceOrInsert(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10))
	depth.Set(item_types.New(200, 20))
	depth.Set(item_types.New(300, 30))
	depth.Set(item_types.New(400, 40))
	depth.Set(item_types.New(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())

	assert.Equal(t, item_types.QuantityType(10), depth.Get(item_types.New(100)).GetQuantity())
	depth.Get(item_types.New(100)).SetQuantity(200)
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.QuantityType(200), depth.Get(item_types.New(100)).GetQuantity())

	item := depth.Get(item_types.New(100))
	item.SetPrice(600)
	depth.Delete(item)
	depth.Set(item_types.New(item.GetPrice(), item.GetQuantity()))
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(100))).GetPrice())
	assert.Equal(t, item_types.PriceType(600), (depth.Get(item_types.New(600))).GetPrice())
	assert.Equal(t, item_types.QuantityType(200), (depth.Get(item_types.New(600))).GetQuantity())
}

func TestGetAndSetDepths(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10))
	depth.Set(item_types.New(200, 20))
	depth.Set(item_types.New(300, 30))
	depth.Set(item_types.New(400, 40))
	depth.Set(item_types.New(500, 50))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())

	otherDepth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	otherDepth.SetTree(depth.GetTree())

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.QuantityType(150), depth.GetSummaQuantity())
	assert.Equal(t, item_types.PriceType(100), depth.Get(item_types.New(100)).GetPrice())
	assert.Equal(t, item_types.PriceType(0), (depth.Get(item_types.New(600))).GetPrice())
}

func TestGetMaxAndSummaValueByPrice(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10))
	depth.Set(item_types.New(200, 20))
	depth.Set(item_types.New(300, 30))
	depth.Set(item_types.New(400, 20))
	depth.Set(item_types.New(500, 10))

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.ValueType(27000), depth.GetSummaValue())

	item, summaValue, _ := depth.GetMaxAndSummaValueByPrice(100, depths_types.UP)
	assert.Equal(t, item_types.PriceType(100), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(1000), item.GetValue())
	assert.Equal(t, item_types.ValueType(1000), summaValue)

	item, summaValue, _ = depth.GetMaxAndSummaValueByPrice(300, depths_types.UP)
	assert.Equal(t, item_types.PriceType(300), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(30), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(9000), item.GetValue())
	assert.Equal(t, item_types.ValueType(14000), summaValue)

	item, summaValue, _ = depth.GetMaxAndSummaValueByPrice(500, depths_types.UP)
	assert.Equal(t, item_types.PriceType(500), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(5000), item.GetValue())
	assert.Equal(t, item_types.ValueType(27000), summaValue)

	item, summaValue, _ = depth.GetMaxAndSummaValueByPrice(100, depths_types.DOWN)
	assert.Equal(t, item_types.PriceType(100), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(1000), item.GetValue())
	assert.Equal(t, item_types.ValueType(27000), summaValue)

	item, summaValue, _ = depth.GetMaxAndSummaValueByPrice(300, depths_types.DOWN)
	assert.Equal(t, item_types.PriceType(300), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(30), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(22000), summaValue)

	item, summaValue, _ = depth.GetMaxAndSummaValueByPrice(500, depths_types.DOWN)
	assert.Equal(t, item_types.PriceType(500), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(5000), item.GetValue())
	assert.Equal(t, item_types.ValueType(5000), summaValue)
}

func TestGetMaxAndSummaValueByPercent(t *testing.T) {
	// TODO: Add test cases.
	depth := depths_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	depth.Set(item_types.New(100, 10)) // 1000 - 1000 - 1000 - 27000
	depth.Set(item_types.New(200, 20)) // 4000 - 5000 - 4000 - 26000
	depth.Set(item_types.New(300, 30)) // 9000 - 14000 - 9000 - 22000
	depth.Set(item_types.New(400, 20)) // 8000 - 22000 - 8000 - 13000
	depth.Set(item_types.New(500, 10)) // 5000 - 27000 - 5000 - 5000

	assert.Equal(t, 5, depth.Count())
	assert.Equal(t, item_types.ValueType(27000), depth.GetSummaValue())

	item, summa, _ := depth.GetMaxAndSummaValueByPercent(25, depths_types.UP) // 27000 * 0.25 = 6750
	assert.Equal(t, item_types.PriceType(200), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(20), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(4000), item.GetValue())
	assert.Equal(t, item_types.ValueType(5000), summa)

	item, summa, _ = depth.GetMaxAndSummaValueByPercent(50, depths_types.UP) // 27000 * 0.5 = 13500
	assert.Equal(t, item_types.PriceType(200), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(20), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(4000), item.GetValue())
	assert.Equal(t, item_types.ValueType(5000), summa)

	item, summa, _ = depth.GetMaxAndSummaValueByPercent(75, depths_types.UP) // 27000 * 0.75 = 20250
	assert.Equal(t, item_types.PriceType(300), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(30), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(9000), item.GetValue())
	assert.Equal(t, item_types.ValueType(14000), summa)

	item, summa, _ = depth.GetMaxAndSummaValueByPercent(75, depths_types.DOWN) // 27000 * 0.75 = 20250
	assert.Equal(t, item_types.PriceType(400), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(20), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(8000), item.GetValue())
	assert.Equal(t, item_types.ValueType(13000), summa)

	item, summa, _ = depth.GetMaxAndSummaValueByPercent(50, depths_types.DOWN) // 27000 * 0.5 = 13500
	assert.Equal(t, item_types.PriceType(400), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(20), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(8000), item.GetValue())
	assert.Equal(t, item_types.ValueType(13000), summa)

	item, summa, _ = depth.GetMaxAndSummaValueByPercent(25, depths_types.DOWN) // 27000 * 0.25 = 6750
	assert.Equal(t, item_types.PriceType(500), item.GetPrice())
	assert.Equal(t, item_types.QuantityType(10), item.GetQuantity())
	assert.Equal(t, item_types.ValueType(5000), item.GetValue())
	assert.Equal(t, item_types.ValueType(5000), summa)
}
