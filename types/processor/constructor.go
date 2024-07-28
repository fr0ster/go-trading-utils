package processor

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	utils "github.com/fr0ster/go-trading-utils/utils"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
)

func New(
	stop chan struct{},
	symbol string,
	exchangeInfo *exchange_types.ExchangeInfo,
	depthsCreator func(*Processor) DepthConstructor,
	ordersCreator func(*Processor) OrdersConstructor,

	getBaseBalance GetBaseBalanceFunction,
	getTargetBalance GetTargetBalanceFunction,
	getFreeBalance GetFreeBalanceFunction,
	getLockedBalance GetLockedBalanceFunction,
	getCurrentPrice GetCurrentPriceFunction,

	getPositionRisk func(*Processor) GetPositionRiskFunction,
	getLeverage GetLeverageFunction,
	setLeverage func(*Processor) SetLeverageFunction,

	getMarginType GetMarginTypeFunction,
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
	symbolInfo := exchangeInfo.GetSymbols().GetSymbol(symbol)
	pp = &Processor{
		exchangeInfo: exchangeInfo,
		symbolInfo:   symbolInfo,
		symbol:       symbol,

		stop:       stop,
		orderTypes: nil,
		degree:     3,
		timeOut:    1 * time.Hour,
		depths:     nil,
		orders:     nil,
	}

	// Налаштовуємо функції
	pp.SetGetterBaseBalanceFunction(getBaseBalance)
	pp.SetGetterTargetBalanceFunction(getTargetBalance)
	pp.SetGetterFreeBalanceFunction(getFreeBalance)
	pp.SetGetterLockedBalanceFunction(getLockedBalance)
	pp.SetGetterCurrentPriceFunction(getCurrentPrice)
	pp.SetGetterPositionRiskFunction(getPositionRisk)
	// Leverage
	pp.SetGetterLeverageFunction(getLeverage)
	pp.SetSetterLeverageFunction(setLeverage)
	// MarginType
	pp.SetGetterMarginTypeFunction(getMarginType)
	pp.SetSetterMarginTypeFunction(setMarginType)
	// PositionMargin
	pp.SetSetterPositionMarginFunction(setPositionMargin)
	// ClosePosition
	pp.SetClosePositionFunction(closePosition)
	// DeltaPrice
	pp.SetGetterDeltaPriceFunction(getDeltaPrice)
	// DeltaQuantity
	pp.SetGetterDeltaQuantityFunction(getDeltaQuantity)
	// LimitOnPosition
	pp.SetGetterLimitOnPositionFunction(getLimitOnPosition)
	// LimitOnTransaction
	pp.SetGetterLimitOnTransactionFunction(getLimitOnTransaction)
	// UpAndLowBound
	pp.SetGetterUpAndLowBoundFunction(getUpAndLowBound)
	// CallbackRate
	pp.SetGetterCallbackRateFunction(getCallbackRate)

	func() {
		price := pp.GetCurrentPrice()
		leverage := pp.GetLeverage()
		transaction := pp.GetLimitOnTransaction()
		_, minLoss := pp.MinPossibleLoss(pp.GetCurrentPrice(), items_types.PriceType(pp.GetUpAndLowBound()), leverage)

		if transaction < minLoss {
			err = fmt.Errorf("limit on transaction %f with price %f isn't enough for open position with leverage %d, we need at least %f or decrease leverage",
				transaction, price, leverage, minLoss)
			return
		}
	}()

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

	// Ініціалізуємо об'єкт
	if depthsCreator != nil {
		pp.depths = depthsCreator(pp)()
	}

	if ordersCreator != nil {
		pp.orders = ordersCreator(pp)()
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

// Налаштовуємо функції
func (pp *Processor) SetGetterBaseBalanceFunction(function GetBaseBalanceFunction) {
	if function != nil {
		pp.getBaseBalance = function
	}
}
func (pp *Processor) SetGetterTargetBalanceFunction(function GetTargetBalanceFunction) {
	if function != nil {
		pp.getTargetBalance = function
	}
}
func (pp *Processor) SetGetterFreeBalanceFunction(function GetFreeBalanceFunction) {
	if function != nil {
		pp.getFreeBalance = function
	}
}
func (pp *Processor) SetGetterLockedBalanceFunction(function GetLockedBalanceFunction) {
	if function != nil {
		pp.getLockedBalance = function
	}
}
func (pp *Processor) SetGetterCurrentPriceFunction(function GetCurrentPriceFunction) {
	if function != nil {
		pp.getCurrentPrice = function
	}
}
func (pp *Processor) SetGetterPositionRiskFunction(function func(*Processor) GetPositionRiskFunction) {
	if function != nil {
		pp.getPositionRisk = function(pp)
	}
}
func (pp *Processor) SetGetterLeverageFunction(function GetLeverageFunction) {
	if function != nil {
		pp.getLeverage = function
	}
}
func (pp *Processor) SetSetterLeverageFunction(function func(*Processor) SetLeverageFunction) {
	if function != nil {
		pp.setLeverage = function(pp)
	}
}
func (pp *Processor) SetGetterMarginTypeFunction(function GetMarginTypeFunction) {
	if function != nil {
		pp.getMarginType = function
	}
}
func (pp *Processor) SetSetterMarginTypeFunction(function func(*Processor) SetMarginTypeFunction) {
	if function != nil {
		pp.setMarginType = function(pp)
	}
}
func (pp *Processor) SetSetterPositionMarginFunction(function func(*Processor) SetPositionMarginFunction) {
	if function != nil {
		pp.setPositionMargin = function(pp)
	}
}
func (pp *Processor) SetClosePositionFunction(function func(*Processor) ClosePositionFunction) {
	if function != nil {
		pp.closePosition = function(pp)
	}
}
func (pp *Processor) SetGetterDeltaPriceFunction(function GetDeltaPriceFunction) {
	if function != nil {
		pp.getDeltaPrice = function
	}
}
func (pp *Processor) SetGetterDeltaQuantityFunction(function GetDeltaQuantityFunction) {
	if function != nil {
		pp.getDeltaQuantity = function
	}
}
func (pp *Processor) SetGetterLimitOnPositionFunction(function GetLimitOnPositionFunction) {
	if function != nil {
		pp.getLimitOnPosition = function
	}
}
func (pp *Processor) SetGetterLimitOnTransactionFunction(function GetLimitOnTransactionFunction) {
	if function != nil {
		pp.getLimitOnTransaction = function
	}
}
func (pp *Processor) SetGetterUpAndLowBoundFunction(function GetUpAndLowBoundFunction) {
	if function != nil {
		pp.getUpAndLowBound = function
	}
}
func (pp *Processor) SetGetterCallbackRateFunction(function GetCallbackRateFunction) {
	if function != nil {
		pp.getCallbackRate = function
	}
}
