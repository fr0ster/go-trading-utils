package processor

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sirupsen/logrus"

	spot_exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

func NewPairProcessor(
	stop chan struct{},
	client *binance.Client,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpBound items_types.PricePercentType,
	LowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	targetPercent items_types.PricePercentType,
	callbackRate items_types.PricePercentType,
	depths ...*depth_types.Depths) (pp *PairProcessor, err error) {
	var (
		depth *depth_types.Depths
	)
	if len(depths) > 0 {
		depth = depths[0]
	}

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

		depth: depth,
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
