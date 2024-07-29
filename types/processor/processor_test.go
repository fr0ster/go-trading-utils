package processor_test

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	"github.com/fr0ster/go-trading-utils/types"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	depths_types "github.com/fr0ster/go-trading-utils/types/depths/depths"
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
	symbol string,
	baseSymbol string,
	targetSymbol string,
	baseBalance items_types.ValueType,
	targetBalance items_types.QuantityType,
	price items_types.PriceType,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	upAndLowBound items_types.PricePercentType,
	notional items_types.ValueType,
	stepSize items_types.QuantityType,
	tickSize items_types.PriceType) (pp *processor.Processor, err error) {
	getSymbols := func() (symbols []*symbol_types.Symbol) {
		symbols = append(symbols, symbol_types.New(
			symbol,                               // symbol
			notional,                             // notional
			stepSize,                             // stepSize
			1000000,                              // maxQty
			0.1,                                  // minQty
			tickSize,                             // tickSize
			100000,                               // maxPrice
			100,                                  // minPrice
			symbol_types.QuoteAsset(baseSymbol),  // quoteAsset
			symbol_types.BaseAsset(targetSymbol), // baseAsset
			false,                                // isMarginTradingAllowed
			nil,                                  // permissions
			nil,                                  // orderType
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
		symbol,       // symbol
		exchangeInfo, // exchangeInfo
		nil,          // depths
		nil,          // orders
		func() items_types.ValueType { return baseBalance },       // getBaseBalance
		func() items_types.QuantityType { return targetBalance },  // getTargetBalance
		func() items_types.ValueType { return baseBalance * 0.5 }, // getFreeBalance
		func() items_types.ValueType { return baseBalance * 0.5 }, // getLockedBalance
		func() items_types.PriceType { return price },             // getCurrentPrice
		nil, // getPositionRisk func(*Processor) GetPositionRiskFunction,

		nil, // getLeverage func(*Processor) GetLeverageFunction,
		nil, // setLeverage func(*Processor) SetLeverageFunction,

		nil, // getMarginType func(*Processor) GetMarginTypeFunction,
		nil, // setMarginType func(*Processor) SetMarginTypeFunction,

		nil, // setPositionMargin func(*Processor) SetPositionMarginFunction,

		nil, // closePosition func(*Processor) ClosePositionFunction,

		nil, // getDeltaPrice GetDeltaPriceFunction,
		nil, // getDeltaQuantity GetDeltaQuantityFunction,

		func() items_types.ValueType { return limitOnPosition },           // getLimitOnPosition
		func() items_types.ValuePercentType { return limitOnTransaction }, // getLimitOnTransaction
		func() items_types.PricePercentType { return upAndLowBound },      // getUpAndLowBound

		nil,  // getCallbackRate GetCallbackRateFunction,
		true, // debug
	)
	return
}

func TestNewSpot(t *testing.T) {
	symbol := "BTCUSDT"
	baseSymbol := "BTC"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(67000)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(100)
	stepSize := items_types.QuantityType(0.001)
	tickSize := items_types.PriceType(0.1)
	maintainer, err := getSpotProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
	assert.Equal(t, "BTCUSDT", maintainer.GetSymbol())
	assert.Equal(t, items_types.ValueType(baseBalance), maintainer.GetBaseBalance())
	assert.Equal(t, items_types.QuantityType(targetBalance), maintainer.GetTargetBalance())
	assert.Equal(t, items_types.ValueType(baseBalance*0.5), maintainer.GetFreeBalance())
	assert.Equal(t, items_types.ValueType(baseBalance*0.5), maintainer.GetLockedBalance())
	assert.Equal(t, items_types.PriceType(price), maintainer.GetCurrentPrice())
	assert.Equal(t, items_types.ValueType(notional), maintainer.GetNotional())
	assert.Equal(t, items_types.QuantityType(0.1), maintainer.GetMinQty())
	assert.Equal(t, items_types.QuantityType(1000000), maintainer.GetMaxQty())
	assert.Equal(t, items_types.PriceType(100), maintainer.GetMinPrice())
	assert.Equal(t, items_types.PriceType(100000), maintainer.GetMaxPrice())
}

func getFuturesProcessor(
	symbol string,
	baseSymbol string,
	targetSymbol string,
	baseBalance items_types.ValueType,
	targetBalance items_types.QuantityType,
	price items_types.PriceType,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	upAndLowBound items_types.PricePercentType,
	notional items_types.ValueType,
	stepSize items_types.QuantityType,
	tickSize items_types.PriceType,
	leverage int) (pp *processor.Processor, err error) {
	getSymbols := func() (symbols []*symbol_types.Symbol) {
		symbols = append(symbols, symbol_types.New(
			symbol,                               // symbol
			notional,                             // notional
			stepSize,                             // stepSize
			1000000,                              // maxQty
			0.1,                                  // minQty
			tickSize,                             // tickSize
			100000,                               // maxPrice
			100,                                  // minPrice
			symbol_types.QuoteAsset(baseSymbol),  // quoteAsset
			symbol_types.BaseAsset(targetSymbol), // baseAsset
			false,                                // isMarginTradingAllowed
			nil,                                  // permissions
			nil,                                  // orderType
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
		symbol,       // symbol
		exchangeInfo, // exchangeInfo
		nil,          // depths
		nil,          // orders
		func() items_types.ValueType {
			return baseBalance
		}, // getBaseBalance
		func() items_types.QuantityType {
			return targetBalance
		}, // getTargetBalance
		func() items_types.ValueType {
			return baseBalance * 0.5
		}, // getFreeBalance
		func() items_types.ValueType {
			return baseBalance * 0.5
		}, // getLockedBalance
		func() items_types.PriceType {
			return price
		}, // getCurrentPrice
		nil, // getPositionRisk
		func() int {
			return leverage
		}, // getLeverage
		nil,           // setLeverage
		nil,           // getMarginType
		nil,           // setMarginType
		nil,           // setPositionMargin
		closePosition, // closePosition
		nil,           // getDeltaPrice
		nil,           // getDeltaQuantity
		func() items_types.ValueType {
			return limitOnPosition
		}, // getLimitOnPosition
		func() items_types.ValuePercentType {
			return limitOnTransaction
		}, // getLimitOnTransaction
		func() items_types.PricePercentType {
			return upAndLowBound
		}, // getUpAndLowBound
		nil,  // getCallbackRate
		true, // debug
	)
	return
}

func TestNewFutures(t *testing.T) {
	symbol := "BTCUSDT"
	baseSymbol := "BTC"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(67000)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(100)
	stepSize := items_types.QuantityType(0.001)
	tickSize := items_types.PriceType(0.1)
	leverage := 10
	maintainer, err := getFuturesProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize,
		leverage)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
	assert.Equal(t, "BTCUSDT", maintainer.GetSymbol())
	assert.Equal(t, items_types.ValueType(baseBalance), maintainer.GetBaseBalance())
	assert.Equal(t, items_types.QuantityType(targetBalance), maintainer.GetTargetBalance())
	assert.Equal(t, items_types.ValueType(baseBalance*0.5), maintainer.GetFreeBalance())
	assert.Equal(t, items_types.ValueType(baseBalance*0.5), maintainer.GetLockedBalance())
	assert.Equal(t, items_types.PriceType(price), maintainer.GetCurrentPrice())
	assert.Equal(t, items_types.ValueType(notional), maintainer.GetNotional())
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

	symbol := "BTCUSDT"
	baseSymbol := "BTC"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(67000)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(100)
	stepSize := items_types.QuantityType(0.001)
	tickSize := items_types.PriceType(0.1)
	pp, err := getSpotProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize)
	assert.Nil(t, err)
	for _, price := range prices {
		assert.Equal(t, price.result, pp.RoundPrice(price.price))
	}
}

func TestGetQuantityByUPnL(t *testing.T) {
	symbol := "CYBERUSDT"
	baseSymbol := "CYBER"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(5)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(5)
	stepSize := items_types.QuantityType(0.1)
	tickSize := items_types.PriceType(100)
	leverage := 10
	pp, err := getFuturesProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize,
		leverage)
	assert.Nil(t, err)
	risk := &futures.PositionRisk{}
	currentPrice := items_types.PriceType(5.0)
	targetOfLoss := items_types.ValueType(200)
	pp.SetGetterLimitOnPositionFunction(func() items_types.ValueType { return targetOfLoss })
	pp.SetGetterLimitOnTransactionFunction(func() items_types.ValuePercentType { return 10 })
	modRisk := func(
		pp *processor.Processor,
		price items_types.PriceType,
		riskIn *futures.PositionRisk,
		side string,
		position float64,
		entryPercent items_types.PricePercentType,
		lossPercent items_types.PricePercentType) (risk *futures.PositionRisk) {
		risk = riskIn
		risk.Notional = utils.ConvFloat64ToStrDefault(float64(notional))
		risk.Leverage = utils.ConvFloat64ToStrDefault(float64(leverage))
		risk.BreakEvenPrice = utils.ConvFloat64ToStrDefault(float64(price) * float64(1+entryPercent/100))
		risk.EntryPrice = utils.ConvFloat64ToStrDefault(float64(price) * float64(1+entryPercent/100))
		risk.Symbol = pp.GetSymbol()
		risk.PositionSide = side
		deltaLiquidation := float64(targetOfLoss) / (position * float64(leverage))
		if risk.PositionSide == "LONG" {
			risk.PositionAmt = utils.ConvFloat64ToStrDefault(position)
			risk.LiquidationPrice = utils.ConvFloat64ToStrDefault(float64(price) - deltaLiquidation)
		} else {
			risk.PositionAmt = utils.ConvFloat64ToStrDefault(-position)
			risk.LiquidationPrice = utils.ConvFloat64ToStrDefault(float64(price) + deltaLiquidation)
		}
		risk.UnRealizedProfit = utils.ConvFloat64ToStrDefault(
			float64(-pp.PossibleLoss(
				items_types.QuantityType(position),
				items_types.PriceType(price*items_types.PriceType(lossPercent/100)))))
		return
	}
	riskLong := modRisk(pp, currentPrice, risk, "LONG", 1000, 10, 10)
	quantity, _ := pp.CalcQuantityByUPnL(depths_types.UP, items_types.PriceType(currentPrice), riskLong)
	assert.Equal(t, items_types.QuantityType(40), quantity)
	riskShort := modRisk(pp, currentPrice, risk, "LONG", 1000, -10, 10)
	quantity, _ = pp.CalcQuantityByUPnL(depths_types.DOWN, items_types.PriceType(currentPrice), riskShort)
	assert.Equal(t, items_types.QuantityType(0.0), quantity)

	riskLong = modRisk(pp, currentPrice, risk, "SHORT", 1000, -10, 10)
	quantity, _ = pp.CalcQuantityByUPnL(depths_types.UP, items_types.PriceType(currentPrice), riskLong)
	assert.Equal(t, items_types.QuantityType(0.0), quantity)
	riskShort = modRisk(pp, currentPrice, risk, "SHORT", 1000, 10, 10)
	quantity, _ = pp.CalcQuantityByUPnL(depths_types.DOWN, items_types.PriceType(currentPrice), riskShort)
	assert.Equal(t, items_types.QuantityType(40), quantity)
}

func TestQuantityAndLossCalculation(t *testing.T) {
	symbol := "CYBERUSDT"
	baseSymbol := "CYBER"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(5)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(5)
	stepSize := items_types.QuantityType(0.1)
	tickSize := items_types.PriceType(100)
	leverage := 10
	pp, err := getFuturesProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize,
		leverage)
	assert.Nil(t, err)
	transaction := pp.GetLimitOnTransaction()
	deltaLiquidation := pp.DeltaLiquidation(leverage)
	assert.Equal(t, items_types.PricePercentType(10), deltaLiquidation)
	delta := items_types.PriceType(deltaLiquidation) * price / 100
	assert.Equal(t, items_types.PriceType(0.5), delta)

	quantity := pp.PossibleQuantity(transaction, price, leverage)
	assert.Equal(t, items_types.QuantityType(500), quantity)

	loss := pp.PossibleLoss(quantity, delta)
	assert.Equal(t, transaction, loss)

	// assert.Equal(t, float64(deltaOnQuantity), float64(delta)*float64(quantity))

	minQuantity := pp.PossibleQuantity(notional, price, leverage)
	assert.Equal(t, items_types.QuantityType(10), minQuantity)
	minLoss := pp.PossibleLoss(minQuantity, delta)
	assert.Equal(t, notional, minLoss)

	test := pp.CheckPosition(price)
	assert.Nil(t, test)
}

func TestGetQuantityAndLoss(t *testing.T) {
	symbol := "CYBERUSDT"
	baseSymbol := "CYBER"
	targetSymbol := "USDT"
	baseBalance := items_types.ValueType(10000)
	price := items_types.PriceType(5)
	targetBalance := items_types.QuantityType(baseBalance) / items_types.QuantityType(price)
	limitOnPosition := baseBalance * 0.25
	limitOnTransaction := items_types.ValuePercentType(10)
	upAndLowBound := items_types.PricePercentType(10)
	notional := items_types.ValueType(5)
	stepSize := items_types.QuantityType(0.1)
	tickSize := items_types.PriceType(100)
	leverage := 10
	pp, err := getFuturesProcessor(
		symbol,
		baseSymbol,
		targetSymbol,
		baseBalance,
		targetBalance,
		price,
		limitOnPosition,
		limitOnTransaction,
		upAndLowBound,
		notional,
		stepSize,
		tickSize,
		leverage)
	assert.Nil(t, err)
	value := items_types.ValueType(5.0)
	quantity := pp.PossibleQuantity(value, price, leverage)
	assert.Equal(t, items_types.QuantityType(10), quantity)
	deltaLiquidation := pp.DeltaLiquidation(leverage)
	assert.Equal(t, items_types.PricePercentType(10), deltaLiquidation)
	loss := pp.PossibleLoss(quantity, price*items_types.PriceType(deltaLiquidation/100))
	assert.Equal(t, items_types.ValueType(5), loss)
}
