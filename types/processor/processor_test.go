package processor_test

import (
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	processor "github.com/fr0ster/go-trading-utils/types/processor"

	"github.com/stretchr/testify/assert"
)

const (
	msg             = "Way too many requests; IP(88.238.33.41) banned until 1721393459728. Please use the websocket for live updates to avoid bans."
	degree          = 3
	timeOut         = 1000 * time.Millisecond
	percentToTarget = 10
	expBase         = 2
)

var (
	quit = make(chan struct{})
)

func TestParserStr(t *testing.T) {
	ip, time, err := processor.ParserErr1003(msg)
	assert.Nil(t, err)
	assert.Equal(t, "88.238.33.41", ip)
	assert.Equal(t, "1721393459728", time)
}

func TestGetLimitPrices(t *testing.T) {
	// TODO: Add test cases.
	bids := bids_types.New(degree, "BTCUSDT")
	bids.Set(items_types.NewBid(100, 10))
	bids.Set(items_types.NewBid(200, 20))
	bids.Set(items_types.NewBid(300, 30))
	bids.Set(items_types.NewBid(400, 20))
	bids.Set(items_types.NewBid(500, 10))
	asks := asks_types.New(degree, "BTCUSDT")
	asks.Set(items_types.NewAsk(600, 10))
	asks.Set(items_types.NewAsk(700, 20))
	asks.Set(items_types.NewAsk(800, 30))
	asks.Set(items_types.NewAsk(900, 20))
	asks.Set(items_types.NewAsk(1000, 10))

	priceTargetUp := items_types.PriceType(700)
	priceTargetDown := items_types.PriceType(400)

	asksFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() > priceTargetUp
	}
	bidsFilter := func(i *items_types.DepthItem) bool {
		return i.GetPrice() < priceTargetDown
	}
	var (
		askMax *items_types.DepthItem
		bidMax *items_types.DepthItem
	)
	_, askMax = asks.GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax = bids.GetFiltered(bidsFilter).GetMinMaxByValue()
	assert.Equal(t, items_types.PriceType(800), askMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), askMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(24000), askMax.GetValue())
	assert.Equal(t, items_types.PriceType(300), bidMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), bidMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(9000), bidMax.GetValue())

	depth := depth_types.New(degree, "BTCUSDT", nil, nil)
	depth.GetAsks().SetTree(asks.GetTree())
	depth.GetBids().SetTree(bids.GetTree())
	_, askMax = depth.GetAsks().GetFiltered(asksFilter).GetMinMaxByValue()
	_, bidMax = depth.GetBids().GetFiltered(bidsFilter).GetMinMaxByValue()
	assert.Equal(t, items_types.PriceType(800), askMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), askMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(24000), askMax.GetValue())
	assert.Equal(t, items_types.PriceType(300), bidMax.GetPrice())
	assert.Equal(t, items_types.QuantityType(30), bidMax.GetQuantity())
	assert.Equal(t, items_types.ValueType(9000), bidMax.GetValue())
}

func TestNew(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	futures_client := futures.NewClient(api_key, secret_key)
	exchangeInfo := exchange_types.New(futures_exchange_info.InitCreator(degree, futures_client))
	maintainer, err := processor.New(
		quit,         // stop
		"BTCUSDT",    // symbol
		exchangeInfo, // exchangeInfo
		nil,          // depths
		nil,          // orders
		nil,          // getBaseBalance
		nil,          // getTargetBalance
		nil,          // getFreeBalance
		nil,          // getLockedBalance
		nil,          // getCurrentPrice
		nil,          // getSymbolInfo
		nil,          // getPositionRisk
		nil,          // setLeverage
		nil,          // setMarginType
		nil,          // setPositionMargin
		nil,          // closePosition
	)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
}
