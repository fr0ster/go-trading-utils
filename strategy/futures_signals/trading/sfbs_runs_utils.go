package futures_signals

import (
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
)

func openPosition(
	sideUp futures.SideType,
	orderTypeUp futures.OrderType,
	sideDown futures.SideType,
	orderTypeDown futures.OrderType,
	closePositionUp bool,
	reduceOnlyUp bool,
	closePositionDown bool,
	reduceOnlyDown bool,
	quantityUp float64,
	quantityDown float64,
	priceUp float64,
	stopPriceUp float64,
	activationPriceUp float64,
	priceDown float64,
	stopPriceDown float64,
	activationPriceDown float64,
	pairProcessor *processor.PairProcessor) (upOrder, downOrder *futures.CreateOrderResponse, err error) {
	err = pairProcessor.CancelAllOrders()
	if err != nil {
		printError()
		return
	}
	if quantityUp != 0 {
		// Створюємо ордери на продаж
		upOrder, err = pairProcessor.CreateOrder(
			orderTypeUp,
			sideUp,
			futures.TimeInForceTypeGTC,
			quantityUp,
			closePositionUp,
			reduceOnlyUp,
			priceUp,
			stopPriceUp,
			activationPriceUp,
			pairProcessor.GetCallbackRate())
		if err != nil {
			logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
				pairProcessor.GetPair(), sideUp, orderTypeUp, priceUp, quantityUp, pairProcessor.GetCallbackRate())
			printError()
			return
		}
		logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
			pairProcessor.GetPair(), sideUp, orderTypeUp, priceUp, quantityUp, pairProcessor.GetCallbackRate(), upOrder.Status)
	}
	if quantityDown != 0 {
		// Створюємо ордери на купівлю
		downOrder, err = pairProcessor.CreateOrder(
			orderTypeDown,
			sideDown,
			futures.TimeInForceTypeGTC,
			quantityDown,
			closePositionDown,
			reduceOnlyDown,
			priceDown,
			stopPriceDown,
			activationPriceDown,
			pairProcessor.GetCallbackRate())
		if err != nil {
			logrus.Errorf("Futures %s: Couldn't set order side %v type %v on price %v with quantity %v call back rate %v",
				pairProcessor.GetPair(), sideDown, orderTypeDown, priceDown, quantityDown, pairProcessor.GetCallbackRate())
			printError()
			return
		}
		logrus.Debugf("Futures %s: Set order side %v type %v on price %v with quantity %v call back rate %v status %v",
			pairProcessor.GetPair(), sideDown, orderTypeDown, priceDown, quantityDown, pairProcessor.GetCallbackRate(), downOrder.Status)
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
