package processor

import (
	"math"
	"time"

	"github.com/adshao/go-binance/v2"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

func NewPairProcessor(
	stop chan struct{},
	client *binance.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	targetPercent float64,
	limitDepth depth_types.DepthAPILimit,
	expBase int,
	callbackRate float64,
	depth ...*depth_types.Depth) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = spot_exchange_info.Init(exchangeInfo, 3, client, symbol)
	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:       client,
		exchangeInfo: exchangeInfo,

		stop: stop,

		pairInfo:   nil,
		orderTypes: map[binance.OrderType]bool{},
		degree:     3,
		// sleepingTime: 1 * time.Second,
		timeOut: 1 * time.Hour,

		depth: nil,
	}
	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.SpotSymbol{Symbol: symbol}).(*symbol_types.SpotSymbol)

	// Ініціалізуємо типи ордерів які можна використовувати для пари
	pp.orderTypes = make(map[binance.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderTypes {
		pp.orderTypes[binance.OrderType(orderType)] = true
	}

	// Буферизуємо інформацію про символ
	pp.symbol, err = pp.GetSymbol().GetSpotSymbol()
	if err != nil {
		return
	}
	pp.baseSymbol = pp.symbol.QuoteAsset
	pp.targetSymbol = pp.symbol.BaseAsset
	pp.notional = utils.ConvStrToFloat64(pp.symbol.NotionalFilter().MinNotional)
	pp.StepSize = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)
	pp.maxQty = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MaxQuantity)
	pp.minQty = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MinQuantity)
	pp.tickSize = utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)
	pp.maxPrice = utils.ConvStrToFloat64(pp.symbol.PriceFilter().MaxPrice)
	pp.minPrice = utils.ConvStrToFloat64(pp.symbol.PriceFilter().MinPrice)

	pp.depth = depth_types.New(pp.degree, symbol, true, targetPercent, limitDepth, expBase+int(math.Log10(pp.tickSize)))
	if pp.depth != nil {
		pp.DepthEventStart(
			stop,
			pp.GetDepthEventCallBack(pp.depth))
	}

	return
}
