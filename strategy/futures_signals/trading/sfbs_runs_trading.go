package futures_signals

import (
	"fmt"
	"math"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func getCallBackTrading(
	pairProcessor *processor.PairProcessor,
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionTPOrderType futures.OrderType,
	shortPositionSLOrderType futures.OrderType,
	longPositionTPOrderType futures.OrderType,
	longPositionSLOrderType futures.OrderType,
	quit chan struct{}) func(*futures.WsUserDataEvent) {
	var (
		sideUp    futures.SideType
		typeUp    futures.OrderType
		priceUp   float64
		sideDown  futures.SideType
		typeDown  futures.OrderType
		priceDown float64
	)
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			risk, _ := pairProcessor.GetPositionRisk()
			currentPrice, _ := pairProcessor.GetCurrentPrice()
			if (event.OrderTradeUpdate.Type == upPositionNewOrderType ||
				event.OrderTradeUpdate.Type == downPositionNewOrderType) &&
				risk != nil && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
				// Спрацював ордер на відкриття позиції
				logrus.Debugf("Futures %s: Order filled %v type %v on price %v/activation price %v with quantity %v side %v status %s",
					pairProcessor.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.Type,
					event.OrderTradeUpdate.LastFilledPrice,
					event.OrderTradeUpdate.ActivationPrice,
					event.OrderTradeUpdate.AccumulatedFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				pairProcessor.CancelAllOrders()
				if event.OrderTradeUpdate.Side == futures.SideTypeBuy {
					// Відкрили позицію long купівлею, закриваємо її продажем
					sideUp = futures.SideTypeSell
					typeUp = longPositionTPOrderType
					priceUp = math.Max(utils.ConvStrToFloat64(risk.BreakEvenPrice), currentPrice) * (1 + pairProcessor.GetDeltaPrice()*2)
					sideDown = futures.SideTypeSell
					typeDown = longPositionSLOrderType
					priceDown = utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice) * (1 - pairProcessor.GetDeltaPrice())
				} else if event.OrderTradeUpdate.Side == futures.SideTypeSell {
					// Відкрили позицію short продажею, закриваємо її купівлею
					sideUp = futures.SideTypeBuy
					typeUp = shortPositionSLOrderType
					priceUp = utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice) * (1 + pairProcessor.GetDeltaPrice())
					sideDown = futures.SideTypeBuy
					typeDown = shortPositionTPOrderType
					priceDown = math.Max(utils.ConvStrToFloat64(risk.BreakEvenPrice), currentPrice) * (1 - pairProcessor.GetDeltaPrice()*2)
				}
				upOrder, downOrder, err := openPosition(
					sideUp,
					typeUp,
					sideDown,
					typeDown,
					false,
					true,
					false,
					true,
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					utils.ConvStrToFloat64(event.OrderTradeUpdate.AccumulatedFilledQty),
					priceUp,
					priceUp,
					priceUp,
					priceDown,
					priceDown,
					priceDown,
					pairProcessor)
				if err != nil {
					logrus.Errorf("Can't open position: %v", err)
					printError()
					close(quit)
					return
				}
				logrus.Debugf("Futures %s: Open position order %v side %v type %v on price %v quantity %v status %v",
					pairProcessor.GetPair(),
					upOrder.OrderID,
					upOrder.Side,
					upOrder.Type,
					upOrder.Price,
					upOrder.OrigQuantity,
					upOrder.Status)
				logrus.Debugf("Futures %s: Open position order %v side %v type %v on price %v quantity %v status %v",
					pairProcessor.GetPair(),
					downOrder.OrderID,
					downOrder.Side,
					downOrder.Type,
					downOrder.Price,
					downOrder.OrigQuantity,
					downOrder.Status)
			} else if risk == nil || utils.ConvStrToFloat64(risk.PositionAmt) == 0 {
				pairProcessor.CancelAllOrders()
				// Створюємо початкові ордери на продаж та купівлю
				if pairProcessor.GetNotional() > pairProcessor.GetLimitOnTransaction() {
					logrus.Errorf("Notional %v > LimitOnTransaction %v", pairProcessor.GetNotional(), pairProcessor.GetLimitOnTransaction())
					printError()
					close(quit)
					return
				}
				currentPrice, err := pairProcessor.GetCurrentPrice()
				if err != nil {
					logrus.Errorf("Can't get current price: %v", err)
					printError()
					close(quit)
					return
				}
				quantity := pairProcessor.RoundQuantity(
					pairProcessor.RoundQuantity(pairProcessor.GetLimitOnTransaction() *
						float64(pairProcessor.GetLeverage()) / currentPrice))
				_, _, err = openPosition(
					upOrderSideOpen,
					upPositionNewOrderType,
					downOrderSideOpen,
					downPositionNewOrderType,
					false,
					false,
					false,
					false,
					quantity,
					quantity,
					pairProcessor.NextPriceUp(currentPrice),
					pairProcessor.NextPriceUp(currentPrice),
					pairProcessor.NextPriceUp(currentPrice),
					pairProcessor.NextPriceDown(currentPrice),
					pairProcessor.NextPriceDown(currentPrice),
					pairProcessor.NextPriceDown(currentPrice),
					pairProcessor)
				if err != nil {
					logrus.Errorf("Can't open position: %v", err)
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func getErrorHandlingTrading(
	pairProcessor *processor.PairProcessor,
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionTPOrderType futures.OrderType,
	shortPositionSLOrderType futures.OrderType,
	longPositionTPOrderType futures.OrderType,
	longPositionSLOrderType futures.OrderType,
	quit chan struct{}) futures.ErrHandler {
	return func(networkErr error) {
		var (
			upNewSide     futures.SideType
			upNewOrder    futures.OrderType
			downNewSide   futures.SideType
			downNewOrder  futures.OrderType
			initPriceUp   float64
			initPriceDown float64
			quantityUp    float64
			quantityDown  float64
			err           error
		)
		openOrders, _ := pairProcessor.GetOpenOrders()
		if len(openOrders) == 0 {
			logrus.Debugf("Futures %s: Error: %v", pairProcessor.GetPair(), networkErr)
			upNewSide,
				upNewOrder,
				downNewSide,
				downNewOrder,
				initPriceUp,
				quantityUp,
				initPriceDown,
				quantityDown,
				err = initNewTradingPosition(
				upOrderSideOpen,
				upPositionNewOrderType,
				downOrderSideOpen,
				downPositionNewOrderType,
				shortPositionTPOrderType,
				shortPositionSLOrderType,
				longPositionTPOrderType,
				longPositionSLOrderType,
				pairProcessor,
				quit)
			if err != nil {
				printError()
				close(quit)
				return
			}
			// Створюємо початкові ордери на продаж та купівлю
			_, _, err = openPosition(
				upNewSide,     // upNewSide
				upNewOrder,    // upNewOrder
				downNewSide,   // downNewSide
				downNewOrder,  // downNewOrder
				false,         // reduceOnlyUp
				false,         // reduceOnlyDown
				false,         // reduceOnlyUp
				false,         // reduceOnlyDown
				quantityUp,    // quantityUp
				quantityDown,  // quantityDown
				initPriceUp,   // initPriceUp
				initPriceUp,   // initPriceUp
				initPriceUp,   // initPriceUp
				initPriceDown, // initPriceDown
				initPriceDown, // initPriceDown
				initPriceDown, // initPriceDown
				pairProcessor) // pairProcessor
			if err != nil {
				printError()
				close(quit)
				return
			}
		}
	}
}

func initNewTradingPosition(
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionTPOrderType futures.OrderType,
	shortPositionSLOrderType futures.OrderType,
	longPositionTPOrderType futures.OrderType,
	longPositionSLOrderType futures.OrderType,
	pairProcessor *processor.PairProcessor,
	quit chan struct{}) (
	upNewSide futures.SideType,
	upNewOrder futures.OrderType,
	downNewSide futures.SideType,
	downNewOrder futures.OrderType,
	initPriceUp float64,
	initPriceDown float64,
	quantityUp float64,
	quantityDown float64,
	err error) {

	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		printError()
		close(quit)
		return
	}
	if pairProcessor.GetLimitOnTransaction() < pairProcessor.GetNotional() {
		err = fmt.Errorf("limit on transaction %v < notional %v", pairProcessor.GetLimitOnTransaction(), pairProcessor.GetNotional())
		printError()
		close(quit)
		return
	}
	price, err := pairProcessor.GetCurrentPrice()
	if err != nil {
		printError()
		close(quit)
		return
	}
	upNewSide = upOrderSideOpen
	upNewOrder = upPositionNewOrderType
	downNewSide = downOrderSideOpen
	downNewOrder = downPositionNewOrderType
	initPriceUp, quantityUp, initPriceDown, quantityDown, _, _, err = pairProcessor.GetPrices(price, risk, true)
	if err != nil {
		printError()
		close(quit)
		return
	}
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 && utils.ConvStrToFloat64(risk.PositionAmt) > pairProcessor.GetNotional() {
		quantityUp = -utils.ConvStrToFloat64(risk.PositionAmt)
		quantityDown = -utils.ConvStrToFloat64(risk.PositionAmt)
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 && utils.ConvStrToFloat64(risk.PositionAmt) > pairProcessor.GetNotional() {
		quantityUp = utils.ConvStrToFloat64(risk.PositionAmt)
		quantityDown = utils.ConvStrToFloat64(risk.PositionAmt)
	}
	upNewSide, upNewOrder, downNewSide, downNewOrder, err = pairProcessor.GetTPAndSLOrdersSideAndTypes(
		risk,
		upOrderSideOpen,
		upPositionNewOrderType,
		downOrderSideOpen,
		downPositionNewOrderType,
		shortPositionTPOrderType,
		shortPositionSLOrderType,
		longPositionTPOrderType,
		longPositionSLOrderType,
		true)
	if err != nil {
		printError()
		close(quit)
		return
	}
	return
}

func RunFuturesTrading(
	client *futures.Client,
	symbol string,
	degree int,
	limit int,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	callBackRate float64,
	upOrderSideOpen futures.SideType,
	upPositionNewOrderType futures.OrderType,
	downOrderSideOpen futures.SideType,
	downPositionNewOrderType futures.OrderType,
	shortPositionTPOrderType futures.OrderType,
	shortPositionSLOrderType futures.OrderType,
	longPositionTPOrderType futures.OrderType,
	longPositionSLOrderType futures.OrderType,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	var (
		upNewSide     futures.SideType
		upNewOrder    futures.OrderType
		downNewSide   futures.SideType
		downNewOrder  futures.OrderType
		initPriceUp   float64
		initPriceDown float64
		quantityUp    float64
		quantityDown  float64
		pairProcessor *processor.PairProcessor
	)
	defer wg.Done()
	futures.WebsocketKeepalive = true
	// Створюємо обробник пари
	pairProcessor, err = processor.NewPairProcessor(
		quit,
		client,
		symbol,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		minSteps,
		callBackRate,
		progression)
	if err != nil {
		printError()
		return
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pairProcessor.GetPair())
	_, err = pairProcessor.UserDataEventStart(
		quit,
		getCallBackTrading(
			pairProcessor,            // pairProcessor
			upOrderSideOpen,          // upOrderSideOpen
			upPositionNewOrderType,   // upPositionNewOrderType
			downOrderSideOpen,        // downOrderSideOpen
			downPositionNewOrderType, // downPositionNewOrderType
			shortPositionTPOrderType, // shortPositionTPOrderType
			shortPositionSLOrderType, // shortPositionSLOrderType
			longPositionTPOrderType,  // longPositionTPOrderType
			longPositionSLOrderType,  // longPositionSLOrderType
			quit),
		getErrorHandlingTrading(
			pairProcessor,            // pairProcessor
			upOrderSideOpen,          // upOrderSideOpen
			upPositionNewOrderType,   // upPositionNewOrderType
			downOrderSideOpen,        // downOrderSideOpen
			downPositionNewOrderType, // downPositionNewOrderType
			shortPositionTPOrderType, // shortPositionTPOrderType
			shortPositionSLOrderType, // shortPositionSLOrderType
			longPositionTPOrderType,  // longPositionTPOrderType
			longPositionSLOrderType,  // longPositionSLOrderType
			quit))                    // quit
	if err != nil {
		printError()
		return err
	}
	upNewSide,
		upNewOrder,
		downNewSide,
		downNewOrder,
		initPriceUp,
		quantityUp,
		initPriceDown,
		quantityDown,
		err = initNewTradingPosition(
		upOrderSideOpen,
		upPositionNewOrderType,
		downOrderSideOpen,
		downPositionNewOrderType,
		shortPositionTPOrderType,
		shortPositionSLOrderType,
		longPositionTPOrderType,
		longPositionSLOrderType,
		pairProcessor,
		quit)
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(
		upNewSide,
		upNewOrder,
		downNewSide,
		downNewOrder,
		false,
		false,
		false,
		false,
		quantityUp,
		quantityDown,
		initPriceUp,
		initPriceUp,
		initPriceUp,
		initPriceDown,
		initPriceDown,
		initPriceDown,
		pairProcessor)
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pairProcessor.GetPair())
	pairProcessor.CancelAllOrders()
	return nil
}
