package processor_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2/futures"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	asks_types "github.com/fr0ster/go-trading-utils/types/depths/asks"
	bids_types "github.com/fr0ster/go-trading-utils/types/depths/bids"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	processor "github.com/fr0ster/go-trading-utils/types/processor"
	symbol_info_types "github.com/fr0ster/go-trading-utils/types/processor/symbol_info"
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

func TestNewFutures(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	futures.UseTestnet = false
	client := futures.NewClient(api_key, secret_key)
	// exchangeInfo := exchange_types.New(futures_exchange_info.InitCreator(degree, client))
	getSymbolInfo := func(pp *processor.Processor) func() (symbolInfo *symbol_info_types.SymbolInfo) {
		return func() (symbolInfo *symbol_info_types.SymbolInfo) {
			// symbol, _ := exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: pp.GetSymbol()}).(*symbol_info.FuturesSymbol).GetFuturesSymbol()
			return symbol_info_types.New(
				// items_types.ValueType(utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().MaxQuantity)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().MinQuantity)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().MaxPrice)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().MinPrice)),
				items_types.ValueType(100),
				items_types.QuantityType(0.001),
				items_types.QuantityType(1000000),
				items_types.QuantityType(0.1),
				items_types.PriceType(0.1),
				items_types.PriceType(100000),
				items_types.PriceType(100),
			)
		}
	}
	getPositionRisk := func() (risks *futures.PositionRisk) {
		var (
			riskArr []*futures.PositionRisk
			err     error
		)
		riskArr, err = client.NewGetPositionRiskService().Do(context.Background())
		if err != nil {
			return
		}
		risks = riskArr[0]
		return
	}
	setLeverage := func(pp *processor.Processor) func(leverage int) (res *futures.SymbolLeverage, err error) {
		return func(leverage int) (res *futures.SymbolLeverage, err error) {
			return client.NewChangeLeverageService().Symbol(pp.GetSymbol()).Leverage(leverage).Do(context.Background())
		}
	}
	setMarginType := func(pp *processor.Processor) func(marginType pairs_types.MarginType) (err error) {
		return func(marginType pairs_types.MarginType) (err error) {
			return client.
				NewChangeMarginTypeService().
				Symbol(pp.GetSymbol()).
				MarginType(futures.MarginType(marginType)).
				Do(context.Background())
		}
	}
	setPositionMargin := func(pp *processor.Processor) func(amountMargin items_types.ValueType, typeMargin int) (err error) {
		return func(amountMargin items_types.ValueType, typeMargin int) (err error) {
			return client.NewUpdatePositionMarginService().
				Symbol(pp.GetSymbol()).Type(typeMargin).
				Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).Do(context.Background())
		}
	}
	closePosition := func(pp *processor.Processor) func(risk *futures.PositionRisk) error {
		return func(risk *futures.PositionRisk) (err error) {
			if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
				_, err = pp.GetOrders().CreateOrder(
					orders_types.OrderType(futures.OrderTypeTakeProfitMarket),
					orders_types.SideType(futures.SideTypeBuy),
					orders_types.TimeInForceType(futures.TimeInForceTypeGTC),
					0, true, false, 0, 0, 0, 0)
			} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
				_, err = pp.GetOrders().CreateOrder(
					orders_types.OrderType(futures.OrderTypeTakeProfitMarket),
					orders_types.SideType(futures.SideTypeSell),
					orders_types.TimeInForceType(futures.TimeInForceTypeGTC),
					0, true, false, 0, 0, 0, 0)
			}
			return
		}
	}
	maintainer, err := processor.New(
		quit,      // stop
		"BTCUSDT", // symbol
		nil,       // exchangeInfo, // exchangeInfo
		nil,       // depths
		nil,       // orders
		func() items_types.ValueType { return 10000 }, // getBaseBalance
		func() items_types.ValueType { return 10000 }, // getTargetBalance
		func() items_types.ValueType { return 10000 }, // getFreeBalance
		func() items_types.ValueType { return 10000 }, // getLockedBalance
		func() items_types.PriceType { return 67000 }, // getCurrentPrice
		getSymbolInfo,     // getSymbolInfo
		getPositionRisk,   // getPositionRisk
		setLeverage,       // setLeverage
		setMarginType,     // setMarginType
		setPositionMargin, // setPositionMargin
		closePosition,     // closePosition
		true,              // debug
	)
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

func TestNewSpot(t *testing.T) {
	// api_key := os.Getenv("API_KEY")
	// secret_key := os.Getenv("SECRET_KEY")
	// futures.UseTestnet = false
	// client := binance.NewClient(api_key, secret_key)
	// exchangeInfo := exchange_types.New(spot_exchange_info.InitCreator(degree, client))
	getSymbolInfo := func(pp *processor.Processor) func() (symbolInfo *symbol_info_types.SymbolInfo) {
		return func() (symbolInfo *symbol_info_types.SymbolInfo) {
			// symbol, _ := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: pp.GetSymbol()}).(*symbol_info.SpotSymbol).GetSpotSymbol()
			return symbol_info_types.New(
				// items_types.ValueType(utils.ConvStrToFloat64(symbol.NotionalFilter().MinNotional)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().MaxQuantity)),
				// items_types.QuantityType(utils.ConvStrToFloat64(symbol.LotSizeFilter().MinQuantity)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().MaxPrice)),
				// items_types.PriceType(utils.ConvStrToFloat64(symbol.PriceFilter().MinPrice)),
				items_types.ValueType(100),
				items_types.QuantityType(0.001),
				items_types.QuantityType(1000000),
				items_types.QuantityType(0.1),
				items_types.PriceType(0.1),
				items_types.PriceType(100000),
				items_types.PriceType(100),
			)
		}
	}
	maintainer, err := processor.New(
		quit,      // stop
		"BTCUSDT", // symbol
		nil,       //exchangeInfo, // exchangeInfo
		nil,       // depths
		nil,       // orders
		func() items_types.ValueType { return 10000 }, // getBaseBalance
		func() items_types.ValueType { return 10000 }, // getTargetBalance
		func() items_types.ValueType { return 10000 }, // getFreeBalance
		func() items_types.ValueType { return 10000 }, // getLockedBalance
		func() items_types.PriceType { return 67000 }, // getCurrentPrice
		getSymbolInfo, // getSymbolInfo
		nil,           // getPositionRisk
		nil,           // setLeverage
		nil,           // setMarginType
		nil,           // setPositionMargin
		nil,           // closePosition
		true,          // debug
	)
	assert.Nil(t, err)
	assert.NotNil(t, maintainer)
	assert.Equal(t, "BTCUSDT", maintainer.GetSymbol())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetBaseBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetTargetBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetFreeBalance())
	assert.Equal(t, items_types.ValueType(10000), maintainer.GetLockedBalance())
	assert.Equal(t, items_types.PriceType(67000), maintainer.GetCurrentPrice())
}
