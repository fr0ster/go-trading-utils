package processor

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	utils "github.com/fr0ster/go-trading-utils/utils"

	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	orders_types "github.com/fr0ster/go-trading-utils/types/orders"
)

func New(
	stop chan struct{},
	symbol string,
	exchangeInfo *exchange_types.ExchangeInfo,
	depths *depth_types.Depths,
	orders *orders_types.Orders,

	getBaseBalance GetBaseBalanceFunction,
	getTargetBalance GetTargetBalanceFunction,
	getFreeBalance GetFreeBalanceFunction,
	getLockedBalance GetLockedBalanceFunction,
	getCurrentPrice GetCurrentPriceFunction,

	getPositionRisk func(*Processor) GetPositionRiskFunction,
	setLeverage func(*Processor) SetLeverageFunction,
	setMarginType func(*Processor) SetMarginTypeFunction,
	setPositionMargin func(*Processor) SetPositionMarginFunction,

	closePosition func(*Processor) ClosePositionFunction,

	getDeltaPrice GetDeltaPriceFunction,
	getDeltaQuantity GetDeltaQuantityFunction,
	getLimitOnPosition GetLimitOnPositionFunction,
	getLimitOnTransaction GetLimitOnTransactionFunction,
	getUpAndLowBound GetUpAndLowBoundFunction,

	getCallbackRate GetCallbackRateFunction,

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
	if getPositionRisk != nil {
		pp.getPositionRisk = getPositionRisk(pp)
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
	if getDeltaPrice != nil {
		pp.getDeltaPrice = getDeltaPrice
	}
	if getDeltaQuantity != nil {
		pp.getDeltaQuantity = getDeltaQuantity
	}
	if getLimitOnTransaction != nil {
		pp.getLimitOnTransaction = getLimitOnTransaction
	}
	if getUpAndLowBound != nil {
		pp.getUpAndLowBound = getUpAndLowBound
	}
	if getCallbackRate != nil {
		pp.getCallbackRate = getCallbackRate
	}

	if len(debug) == 0 || (len(debug) > 0 && !debug[0]) {
		if pp.setLeverage != nil {
			leverage, _, _, _ := pp.SetLeverage(pp.GetLeverage())
			if leverage != pp.GetLeverage() {
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
