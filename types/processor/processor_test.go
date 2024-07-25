package processor_test

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/fr0ster/go-trading-utils/types"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	processor "github.com/fr0ster/go-trading-utils/types/processor"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	symbols_types "github.com/fr0ster/go-trading-utils/types/symbols"
	utils "github.com/fr0ster/go-trading-utils/utils"

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

func getSpotProcessor(
	notional items_types.ValueType,
	stepSize items_types.QuantityType,
	tickSize items_types.PriceType) (pp *processor.Processor, err error) {
	getSymbols := func() (symbols []*symbol_types.Symbol) {
		symbols = append(symbols, symbol_types.New(
			"BTCUSDT", // symbol
			notional,  // notional
			stepSize,  // stepSize
			1000000,   // maxQty
			0.1,       // minQty
			tickSize,  // tickSize
			100000,    // maxPrice
			100,       // minPrice
			"USDT",    // quoteAsset
			"BTC",     // baseAsset
			false,     // isMarginTradingAllowed
			nil,       // permissions
			nil,       // orderType
		))
		return
	}
	init := func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
			val.Timezone = time.Now().Location().String()
			val.ServerTime = time.Now().Unix()
			val.RateLimits = nil
			val.ExchangeFilters = nil
			val.Symbols, _ = symbols_types.New(degree, getSymbols)
			return
		}
	}
	exchangeInfo := exchange_types.New(init)
	pp, err = processor.New(
		quit,         // stop
		"BTCUSDT",    // symbol
		exchangeInfo, // exchangeInfo
		nil,          // depths
		nil,          // orders
		func() items_types.ValueType { return 10000 }, // getBaseBalance
		func() items_types.ValueType { return 10000 }, // getTargetBalance
		func() items_types.ValueType { return 10000 }, // getFreeBalance
		func() items_types.ValueType { return 10000 }, // getLockedBalance
		func() items_types.PriceType { return 67000 }, // getCurrentPrice
		nil, // getPositionRisk func(*Processor) GetPositionRiskFunction,
		nil, // setLeverage func(*Processor) SetLeverageFunction,
		nil, // setMarginType func(*Processor) SetMarginTypeFunction,
		nil, // setPositionMargin func(*Processor) SetPositionMarginFunction,

		nil, // closePosition func(*Processor) ClosePositionFunction,

		nil, // getDeltaPrice GetDeltaPriceFunction,
		nil, // getDeltaQuantity GetDeltaQuantityFunction,
		nil, // getLimitOnPosition GetLimitOnPositionFunction,
		nil, // getLimitOnTransaction GetLimitOnTransactionFunction,
		nil, // getUpAndLowBound GetUpAndLowBoundFunction,

		nil,  // getCallbackRate GetCallbackRateFunction,
		true, // debug
	)
	return
}

func TestNewSpot(t *testing.T) {
	maintainer, err := getSpotProcessor(100, 0.001, 0.1)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
	assert.Equal(t, "BTCUSDT", maintainer.GetSymbol())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetBaseBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetTargetBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetFreeBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetLockedBalance())
	assert.Equal(t, items_types.PriceType(67000), maintainer.GetCurrentPrice())
	assert.Equal(t, items_types.ValueType(100), maintainer.GetNotional())
	assert.Equal(t, items_types.QuantityType(0.1), maintainer.GetMinQty())
	assert.Equal(t, items_types.QuantityType(1000000), maintainer.GetMaxQty())
	assert.Equal(t, items_types.PriceType(100), maintainer.GetMinPrice())
	assert.Equal(t, items_types.PriceType(100000), maintainer.GetMaxPrice())
}

func getFuturesProcessor(
	notional items_types.ValueType,
	stepSize items_types.QuantityType,
	tickSize items_types.PriceType) (pp *processor.Processor, err error) {
	getSymbols := func() (symbols []*symbol_types.Symbol) {
		symbols = append(symbols, symbol_types.New(
			"BTCUSDT", // symbol
			notional,  // notional
			stepSize,  // stepSize
			1000000,   // maxQty
			0.1,       // minQty
			tickSize,  // tickSize
			100000,    // maxPrice
			100,       // minPrice
			"USDT",    // quoteAsset
			"BTC",     // baseAsset
			false,     // isMarginTradingAllowed
			nil,       // permissions
			nil,       // orderType
		))
		return
	}
	init := func(val *exchange_types.ExchangeInfo) types.InitFunction {
		return func() (err error) {
			val.Timezone = time.Now().Location().String()
			val.ServerTime = time.Now().Unix()
			val.RateLimits = nil
			val.ExchangeFilters = nil
			val.Symbols, _ = symbols_types.New(degree, getSymbols)
			return
		}
	}
	exchangeInfo := exchange_types.New(init)
	closePosition := func(pp *processor.Processor) processor.ClosePositionFunction {
		return func() (err error) {
			risk := pp.GetPositionRisk()
			if risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
				if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
					_, err = pp.GetOrders().CreateOrder(
						types.OrderType(futures.OrderTypeTakeProfitMarket),
						types.SideType(futures.SideTypeBuy),
						types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
					_, err = pp.GetOrders().CreateOrder(
						types.OrderType(futures.OrderTypeTakeProfitMarket),
						types.SideType(futures.SideTypeSell),
						types.TimeInForceType(futures.TimeInForceTypeGTC),
						0, true, false, 0, 0, 0, 0)
				}
			}
			return
		}
	}
	pp, err = processor.New(
		quit,         // stop
		"BTCUSDT",    // symbol
		exchangeInfo, // exchangeInfo
		nil,          // depths
		nil,          // orders
		func() items_types.ValueType { return 10000 }, // getBaseBalance
		func() items_types.ValueType { return 10000 }, // getTargetBalance
		func() items_types.ValueType { return 1000 },  // getFreeBalance
		func() items_types.ValueType { return 10000 }, // getLockedBalance
		func() items_types.PriceType { return 67000 }, // getCurrentPrice
		nil,           // getPositionRisk
		nil,           // setLeverage
		nil,           // setMarginType
		nil,           // setPositionMargin
		closePosition, // closePosition
		nil,           // getDeltaPrice
		nil,           // getDeltaQuantity
		nil,           // getLimitOnPosition
		nil,           // getLimitOnTransaction
		nil,           // getUpAndLowBound
		nil,           // getCallbackRate
		true,          // debug
	)
	return
}

func TestNewFutures(t *testing.T) {
	maintainer, err := getFuturesProcessor(100, 0.001, 0.1)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
	assert.Equal(t, "BTCUSDT", maintainer.GetSymbol())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetBaseBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetTargetBalance())
	assert.Equal(t, items_types.ValueType(1000), maintainer.GetFreeBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetLockedBalance())
	assert.Equal(t, items_types.PriceType(67000), maintainer.GetCurrentPrice())
	assert.Equal(t, items_types.ValueType(100), maintainer.GetNotional())
	assert.Equal(t, items_types.QuantityType(0.1), maintainer.GetMinQty())
	assert.Equal(t, items_types.QuantityType(1000000), maintainer.GetMaxQty())
	assert.Equal(t, items_types.PriceType(100), maintainer.GetMinPrice())
	assert.Equal(t, items_types.PriceType(100000), maintainer.GetMaxPrice())
}

func TestRoundPrice(t *testing.T) {
	type pairs struct {
		price  items_types.PriceType
		result items_types.PriceType
	}
	prices := []pairs{
		{67100.11, 67100.1},
		{67100.111, 67100.1},
		{67100.1111, 67100.1},
		{67100.11111, 67100.1},
		{67100.111111, 67100.1},
		{67100.1111111, 67100.1},
	}
	pp, err := getSpotProcessor(100, 0.001, 0.1)
	assert.Nil(t, err)
	for _, price := range prices {
		assert.Equal(t, price.result, pp.RoundPrice(price.price))
	}
}
