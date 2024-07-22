package processor

import (
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	futures_exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	utils "github.com/fr0ster/go-trading-utils/utils"
	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
)

func NewPairProcessor(
	stop chan struct{},
	client *futures.Client,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpBound items_types.PricePercentType,
	LowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	callbackRate items_types.PricePercentType,
	progression pairs_types.ProgressionType,
	depths ...*depth_types.Depths) (pp *PairProcessor, err error) {
	var (
		depth *depth_types.Depths
	)
	if len(depths) > 0 {
		depth = depths[0]
	}

	exchangeInfo := exchange_types.New()
	err = futures_exchange_info.RestrictedInit(exchangeInfo, 3, []string{symbol}, client)
	if err != nil {
		err = ParseError(err)
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
		timeOut:            1 * time.Hour,
		limitOnPosition:    limitOnPosition,
		limitOnTransaction: limitOnTransaction,
		UpBoundPercent:     UpBound,
		LowBoundPercent:    LowBound,
		leverage:           leverage,
		marginType:         marginType,
		callbackRate:       callbackRate,

		deltaPrice:    items_types.PriceType(deltaPrice),
		deltaQuantity: items_types.QuantityType(deltaQuantity),

		progression: progression,
		depth:       depth,
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
	pp.notional = items_types.ValueType(utils.ConvStrToFloat64(pp.symbol.MinNotionalFilter().Notional))
	pp.StepSize = items_types.QuantityType(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize))
	pp.maxQty = items_types.QuantityType(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MaxQuantity))
	pp.minQty = items_types.QuantityType(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().MinQuantity))
	pp.tickSize = items_types.PriceType(utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize))
	pp.maxPrice = items_types.PriceType(utils.ConvStrToFloat64(pp.symbol.PriceFilter().MaxPrice))
	pp.minPrice = items_types.PriceType(utils.ConvStrToFloat64(pp.symbol.PriceFilter().MinPrice))

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
		err = ParseError(err)
		return
	}
	if leverage != 0 {
		_ = pp.SetMarginType(marginType) // Встановлюємо тип маржі, як зміна не потрібна, помилку ігноруємо
		res, _ := pp.SetLeverage(leverage)
		if res.Leverage != leverage {
			err = fmt.Errorf("leverage %v is not supported", leverage)
			err = ParseError(err)
			return
		}
	}

	return
}

func printError() {
	if logrus.GetLevel() == logrus.DebugLevel {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			logrus.Errorf("Error occurred in file: %s at line: %d", file, line)
		} else {
			logrus.Errorf("Error occurred but could not get the caller information")
		}
	}
}

func ParserErr1003(msg string) (ip, time string, err error) {
	re := regexp.MustCompile(`IP\(([\d\.]+)\) banned until (\d+)`)
	matches := re.FindStringSubmatch(msg)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("failed to parse")
	}
	ip, time = matches[1], matches[2]
	return
}

func ParseError(err error) error {
	apiErr, _ := utils.ParseAPIError(err)
	printError()
	switch apiErr.Code {
	case -1003:
		ip, timeStr, err := ParserErr1003(apiErr.Msg)
		timestamp, errParse := strconv.ParseInt(timeStr, 10, 64)
		if errParse != nil {
			return err
		}

		bannedTime := time.UnixMilli(timestamp)
		return fmt.Errorf("way too many requests; IP %s banned until: %s", ip, bannedTime)
	default:
		return err
	}
}
