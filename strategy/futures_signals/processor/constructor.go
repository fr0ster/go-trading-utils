package processor

import (
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"
	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func NewPairProcessor(
	stop chan struct{},
	client *futures.Client,
	symbol string,
	limitOnPosition float64,
	limitOnTransaction float64,
	UpBound float64,
	LowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	callbackRate float64,
	progression pairs_types.ProgressionType,
	depth ...*depth_types.Depth) (pp *PairProcessor, err error) {
	exchangeInfo := exchange_types.New()
	err = futures_exchange_info.Init(exchangeInfo, 3, client)
	if err != nil {
		return
	}
	pp = &PairProcessor{
		client:       client,
		exchangeInfo: exchangeInfo,
		symbol:       nil,
		baseSymbol:   "",
		notional:     0,
		StepSize:     0,
		minSteps:     minSteps,
		up:           btree.New(2),
		down:         btree.New(2),

		stop: stop,

		pairInfo:           nil,
		orderTypes:         nil,
		degree:             3,
		sleepingTime:       1 * time.Second,
		timeOut:            1 * time.Hour,
		limitOnPosition:    limitOnPosition,
		limitOnTransaction: limitOnTransaction,
		UpBoundPercent:     UpBound,
		UpBound:            0,
		LowBoundPercent:    LowBound,
		LowBound:           0,
		leverage:           leverage,
		marginType:         marginType,
		callbackRate:       callbackRate,

		deltaPrice:    deltaPrice,
		deltaQuantity: deltaQuantity,

		progression: progression,

		depth: nil,
	}

	if len(depth) > 0 {
		pp.depth = depth[0]
		if pp.depth != nil {
			pp.DepthEventStart(
				stop,
				pp.depth.GetLimitStream(),
				pp.depth.GetRateStream(),
				pp.GetDepthEventCallBack(pp.depth))
		}
	}

	// Ініціалізуємо інформацію про пару
	pp.pairInfo = pp.exchangeInfo.GetSymbol(
		&symbol_types.FuturesSymbol{Symbol: symbol}).(*symbol_types.FuturesSymbol)

	// Ініціалізуємо типи ордерів
	pp.orderTypes = make(map[futures.OrderType]bool, 0)
	for _, orderType := range pp.pairInfo.OrderType {
		pp.orderTypes[orderType] = true
	}

	// Буферизуємо інформацію про символ
	pp.symbol, err = pp.GetSymbol().GetFuturesSymbol()
	if err != nil {
		return
	}
	pp.baseSymbol = pp.symbol.QuoteAsset
	pp.targetSymbol = pp.symbol.BaseAsset
	pp.notional = utils.ConvStrToFloat64(pp.symbol.MinNotionalFilter().Notional)
	pp.StepSize = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)
	pp.maxQty = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MaxQuantity)
	pp.minQty = utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MinQuantity)
	pp.tickSize = utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)
	pp.maxPrice = utils.ConvStrToFloat64(pp.symbol.PriceFilter().MaxPrice)
	pp.minPrice = utils.ConvStrToFloat64(pp.symbol.PriceFilter().MinPrice)

	currentPrice, err := pp.GetCurrentPrice()
	if err != nil {
		return
	}
	pp.SetBounds(currentPrice)

	if pp.progression == pairs_types.ArithmeticProgression {
		pp.NthTerm = progressions.ArithmeticProgressionNthTerm
		pp.Sum = progressions.ArithmeticProgressionSum
		pp.FindNthTerm = progressions.FindArithmeticProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfArithmeticProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 - P1 }
		pp.FindProgressionTthTerm = progressions.FindArithmeticProgressionTthTerm
	} else if pp.progression == pairs_types.GeometricProgression {
		pp.NthTerm = progressions.GeometricProgressionNthTerm
		pp.Sum = progressions.GeometricProgressionSum
		pp.FindNthTerm = progressions.FindGeometricProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfGeometricProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 / P1 }
		pp.FindProgressionTthTerm = progressions.FindGeometricProgressionTthTerm
	} else if pp.progression == pairs_types.CubicProgression {
		pp.NthTerm = progressions.CubicProgressionNthTerm
		pp.Sum = progressions.CubicProgressionSum
		pp.FindNthTerm = progressions.FindCubicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfCubicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return math.Pow(P2/P1, 1.0/3) }
		pp.FindProgressionTthTerm = progressions.FindCubicProgressionTthTerm
	} else if pp.progression == pairs_types.CubicRootProgression {
		pp.NthTerm = progressions.CubicRootProgressionNthTerm
		pp.Sum = progressions.CubicRootProgressionSum
		pp.FindNthTerm = progressions.FindCubicRootProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfCubicRootProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return math.Cbrt(P2 / P1) }
		pp.FindProgressionTthTerm = progressions.FindCubicRootProgressionTthTerm
	} else if pp.progression == pairs_types.QuadraticProgression {
		pp.NthTerm = progressions.QuadraticProgressionNthTerm
		pp.Sum = progressions.QuadraticProgressionSum
		pp.FindNthTerm = progressions.FindQuadraticProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfQuadraticProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return (P2 - P1) / 1 }
		pp.FindProgressionTthTerm = progressions.FindQuadraticProgressionTthTerm
	} else if pp.progression == pairs_types.ExponentialProgression {
		pp.NthTerm = progressions.ExponentialProgressionNthTerm
		pp.Sum = progressions.ExponentialProgressionSum
		pp.FindNthTerm = progressions.FindExponentialProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfExponentialProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return P2 / P1 }
		pp.FindProgressionTthTerm = progressions.FindExponentialProgressionTthTerm
	} else if pp.progression == pairs_types.LogarithmicProgression {
		pp.NthTerm = progressions.LogarithmicProgressionNthTerm
		pp.Sum = progressions.LogarithmicProgressionSum
		pp.FindNthTerm = progressions.FindLogarithmicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfLogarithmicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return (P2 - P1) / math.Log(2) }
		pp.FindProgressionTthTerm = progressions.FindLogarithmicProgressionTthTerm
	} else if pp.progression == pairs_types.HarmonicProgression {
		pp.NthTerm = progressions.HarmonicProgressionNthTerm
		pp.Sum = progressions.HarmonicProgressionSum
		pp.FindNthTerm = progressions.FindHarmonicProgressionNthTerm
		pp.FindLengthOfProgression = progressions.FindLengthOfHarmonicProgression
		pp.GetDelta = func(P1, P2 float64) float64 { return 1/P2 - 1/P1 }
		pp.FindProgressionTthTerm = progressions.FindHarmonicProgressionTthTerm
	} else {
		err = fmt.Errorf("progression type %v is not supported", pp.progression)
		return
	}
	_ = pp.SetMarginType(marginType) // Встановлюємо тип маржі, як зміна не потрібна, помилку ігноруємо
	res, _ := pp.SetLeverage(leverage)
	if res.Leverage != leverage {
		err = fmt.Errorf("leverage %v is not supported", leverage)
		return
	}

	return
}
