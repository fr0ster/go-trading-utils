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

	getBaseBalance func() items_types.ValueType,
	getTargetBalance func() items_types.ValueType,
	getFreeBalance func() items_types.ValueType,
	getLockedBalance func() items_types.ValueType,
	getCurrentPrice func() items_types.PriceType,

	// getSymbolInfo func(*Processor) func() *symbol_types.SymbolInfo,
	getPositionRisk func() *futures.PositionRisk,
	setLeverage func(*Processor) func(int) (*futures.SymbolLeverage, error),
	setMarginType func(*Processor) func(pairs_types.MarginType) error,
	setPositionMargin func(*Processor) func(items_types.ValueType, int) error,

	closePosition func(*Processor) func(*futures.PositionRisk) error,
	debug ...bool) (pp *Processor, err error) {
	pp = &Processor{
		exchangeInfo: exchangeInfo,
		symbol:       symbol,

		stop:       stop,
		orderTypes: nil,
		degree:     3,
		timeOut:    1 * time.Hour,
		depths:     depths,
		orders:     orders,
	}

	if getBaseBalance != nil {
		pp.getBaseBalance = getBaseBalance
	}
	if getTargetBalance != nil {
		pp.getTargetBalance = getTargetBalance
	}
	if getFreeBalance != nil {
		pp.getFreeBalance = getFreeBalance
	}
	if getLockedBalance != nil {
		pp.getLockedBalance = getLockedBalance
	}
	if getCurrentPrice != nil {
		pp.getCurrentPrice = getCurrentPrice
	}
	// if getSymbolInfo != nil {
	// 	pp.getSymbolInfo = getSymbolInfo(pp)
	// 	pp.symbolInfo = pp.getSymbolInfo()
	// }
	if getPositionRisk != nil {
		pp.getPositionRisk = getPositionRisk
	}
	if setLeverage != nil {
		pp.setLeverage = setLeverage(pp)
	}
	if setMarginType != nil {
		pp.setMarginType = setMarginType(pp)
	}
	if setPositionMargin != nil {
		pp.setPositionMargin = setPositionMargin(pp)
	}
	if closePosition != nil {
		pp.closePosition = closePosition(pp)
	}
	if len(debug) == 0 || (len(debug) > 0 && !debug[0]) {
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
