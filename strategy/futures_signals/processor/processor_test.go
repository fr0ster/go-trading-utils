package processor_test

import (
	"testing"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	asks_types "github.com/fr0ster/go-trading-utils/types/depth/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depth/bids"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"

	"github.com/stretchr/testify/assert"
)

const (
	msg    = "Way too many requests; IP(88.238.33.41) banned until 1721393459728. Please use the websocket for live updates to avoid bans."
	degree = 3
)

func TestParserStr(t *testing.T) {
	ip, time, err := processor.ParserErr1003(msg)
	assert.Nil(t, err)
	assert.Equal(t, "88.238.33.41", ip)
	assert.Equal(t, "1721393459728", time)
}

func TestGetLimitPrices(t *testing.T) {
	// TODO: Add test cases.
	bids := bids_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	bids.Set(items_types.NewBid(100, 10))
	bids.Set(items_types.NewBid(200, 20))
	bids.Set(items_types.NewBid(300, 30))
	bids.Set(items_types.NewBid(400, 20))
	bids.Set(items_types.NewBid(500, 10))
	asks := asks_types.New(degree, "BTCUSDT", 10, 100, 2, depths_types.DepthStreamRate100ms)
	asks.Set(items_types.NewAsk(600, 10))
	asks.Set(items_types.NewAsk(700, 20))
	asks.Set(items_types.NewAsk(800, 30))
	asks.Set(items_types.NewAsk(900, 20))
	asks.Set(items_types.NewAsk(1000, 10))

	priceTargetDown := items_types.PriceType(400)
	priceTargetUp := items_types.PriceType(700)

	asksFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() > priceTargetUp
	}
	bidsFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() < priceTargetDown
	}
	_, askMax := asks.GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax := bids.GetFiltered(bidsFilter).GetMinMaxByValue()
	assert.Equal(t, items_types.PriceType(800), askMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), askMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(24000), askMax.GetValue())
	assert.Equal(t, items_types.PriceType(300), bidMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), bidMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(9000), bidMax.GetValue())
}
