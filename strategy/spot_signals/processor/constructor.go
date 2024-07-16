package processor

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

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
		err = ParseError(err)
		return
	}
	pp = &PairProcessor{
		client:       client,
		exchangeInfo: exchangeInfo,

		stop: stop,

		pairInfo:   nil,
		orderTypes: map[binance.OrderType]bool{},
		degree:     3,
		timeOut:    1 * time.Hour,

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
		err = ParseError(err)
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

func ParseError(err error) error {
	apiErr, _ := utils.ParseAPIError(err)
	printError()
	switch apiErr.Code {
	case -1003:
		var (
			bannedIP    string
			bannedUntil string
		)
		_, errScanf := fmt.Sscanf(apiErr.Msg, "Way too many requests; IP(%s) banned until %s. Please use the websocket for live updates to avoid bans.",
			&bannedIP, &bannedUntil)
		if errScanf != nil {
			return err
		}
		timestamp, errParse := strconv.ParseInt(bannedUntil, 10, 64)
		if errParse != nil {
			return err
		}

		// Для Go 1.17 і вище
		bannedTime := time.UnixMilli(timestamp)
		return fmt.Errorf("way too many requests; IP banned until: %s", bannedTime)
	default:
		return err
	}
}
