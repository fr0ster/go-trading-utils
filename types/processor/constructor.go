package processor

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	utils "github.com/fr0ster/go-trading-utils/utils"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

func New(
	stop chan struct{},
	symbol string,
	exchangeInfo *exchange_types.ExchangeInfo,
	depths *depth_types.Depths,
	orders *orders_types.Orders,
	getBaseBalance func() (items_types.ValueType, error),
	getTargetBalance func() (items_types.ValueType, error),
	getFreeBalance func() items_types.ValueType,
	getLockedBalance func() (items_types.ValueType, error),
	getCurrentPrice func() (items_types.PriceType, error),
	getSymbolInfo func() (SymbolInfo, error),
	getPositionRisk func() (risks *futures.PositionRisk, err error),
	setLeverage func(leverage int) (res *futures.SymbolLeverage, err error),
	setMarginType func(marginType pairs_types.MarginType) (err error),
	setPositionMargin func(amountMargin items_types.ValueType, typeMargin int) (err error),
	closePosition func(risk *futures.PositionRisk) (err error)) (pp *Processor, err error) {
	pp = &Processor{
		exchangeInfo: exchangeInfo,
		symbol:       symbol,
		symbolInfo:   SymbolInfo{},

		stop:       stop,
		orderTypes: nil,
		degree:     3,
		timeOut:    1 * time.Hour,
		depth:      depths,
		orders:     orders,
	}

	if getBaseBalance != nil {
		pp.GetBaseBalance = getBaseBalance
	}
	if getTargetBalance != nil {
		pp.GetTargetBalance = getTargetBalance
	}
	if getFreeBalance != nil {
		pp.GetFreeBalance = getFreeBalance
	}
	if getLockedBalance != nil {
		pp.GetLockedBalance = getLockedBalance
	}
	if getCurrentPrice != nil {
		pp.GetCurrentPrice = getCurrentPrice
	}
	if getSymbolInfo != nil {
		pp.GetSymbolInfo = getSymbolInfo
	}
	if getPositionRisk != nil {
		pp.getPositionRisk = getPositionRisk
	}
	if setLeverage != nil {
		pp.setLeverage = setLeverage
	}
	if setMarginType != nil {
		pp.setMarginType = setMarginType
	}
	if setPositionMargin != nil {
		pp.setPositionMargin = setPositionMargin
	}
	if closePosition != nil {
		pp.closePosition = closePosition
	}
	if pp.setLeverage != nil {
		res, _ := pp.SetLeverage(pp.GetLeverage())
		if res.Leverage != pp.GetLeverage() {
			err = fmt.Errorf("leverage %v is not supported", pp.GetLeverage())
			err = ParseError(err)
			return
		}
	}
	if pp.setMarginType != nil {
		_ = pp.SetMarginType(pp.GetMarginType()) // Встановлюємо тип маржі, як зміна не потрібна, помилку ігноруємо
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
