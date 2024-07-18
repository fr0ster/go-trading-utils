package grid

import (
	"math"
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types"
	depth_items "github.com/fr0ster/go-trading-utils/types/depth/items"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func initGrid(
	pairProcessor *processor.PairProcessor,
	price depth_items.PriceType,
	quantity depth_items.QuantityType,
	sellOrder,
	buyOrder *futures.CreateOrderResponse) (grid *grid_types.Grid, err error) {
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pairProcessor.GetPair())
	grid = grid_types.New()
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, 0, pairProcessor.RoundPrice(price*(1+pairProcessor.GetDeltaPrice())), pairProcessor.RoundPrice(price*(1-pairProcessor.GetDeltaPrice())), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pairProcessor.GetPair(), price)
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, pairProcessor.RoundPrice(price*(1+pairProcessor.GetDeltaPrice())), quantity, 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pairProcessor.GetPair(), pairProcessor.RoundPrice(price*(1+pairProcessor.GetDeltaPrice())))
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, pairProcessor.RoundPrice(price*(1-pairProcessor.GetDeltaPrice())), quantity, price, 0, types.SideTypeBuy))
	grid.Debug("Futures Grid", "", pairProcessor.GetPair())
	return
}

func initVars(
	pairProcessor *processor.PairProcessor) (
	price depth_items.PriceType,
	priceUp,
	priceDown depth_items.PriceType,
	minNotional float64,
	err error) {
	symbol, err := pairProcessor.GetFuturesSymbol()
	if err != nil {
		return
	}
	// Отримання середньої ціни
	price, _ = pairProcessor.GetCurrentPrice() // Отримання ціни по ринку для пари
	price = pairProcessor.RoundPrice(price)
	priceUp = pairProcessor.NextPriceUp(price)
	priceDown = pairProcessor.NextPriceDown(price)
	minNotional = utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
	return
}

func marginBalancing(
	risk *futures.PositionRisk,
	pairProcessor *processor.PairProcessor) (err error) {
	// Балансування маржі як треба
	if utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		delta := pairProcessor.RoundPrice(pairProcessor.GetFreeBalance()) - pairProcessor.RoundPrice(depth_items.PriceType(utils.ConvStrToFloat64(risk.IsolatedMargin)))
		if delta != 0 {
			if delta > 0 && delta < pairProcessor.GetFreeBalance() {
				err = pairProcessor.SetPositionMargin(delta, 1)
				logrus.Debugf("Futures %s: IsolatedMargin %v < current position balance %v and we have enough free %v",
					pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance(), pairProcessor.GetFreeBalance())
			}
		}
	}
	return
}

func openPosition(
	sideUp futures.SideType,
	orderTypeUp futures.OrderType,
	sideDown futures.SideType,
	orderTypeDown futures.OrderType,
	closePositionUp bool,
	reduceOnlyUp bool,
	closePositionDown bool,
	reduceOnlyDown bool,
	quantityUp depth_items.QuantityType,
	quantityDown depth_items.QuantityType,
	priceUp depth_items.PriceType,
	stopPriceUp depth_items.PriceType,
	activationPriceUp depth_items.PriceType,
	priceDown depth_items.PriceType,
	stopPriceDown depth_items.PriceType,
	activationPriceDown depth_items.PriceType,
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

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	pairProcessor *processor.PairProcessor,
	side futures.SideType,
	grid *grid_types.Grid,
	percentsToStopSettingNewOrder float64,
	order *grid_types.Record,
	quantity depth_items.QuantityType,
	locked depth_items.PriceType,
	risk *futures.PositionRisk) (err error) {
	var (
		takerRecord *grid_types.Record
		takerOrder  *futures.CreateOrderResponse
	)
	delta_percent := func(currentPrice depth_items.PriceType) depth_items.PriceType {
		return depth_items.PriceType(
			math.Abs((float64(currentPrice) - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice)))
	}
	if side == futures.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			upPrice := pairProcessor.RoundPrice(order.GetPrice() * (1 + pairProcessor.GetDeltaPrice()))
			if (pairProcessor.GetUpBound() == 0 || upPrice <= pairProcessor.GetUpBound()) &&
				depth_items.PriceType(utils.ConvStrToFloat64(risk.IsolatedMargin)) <= pairProcessor.GetFreeBalance() &&
				locked <= pairProcessor.GetFreeBalance() {
				upOrder, err := pairProcessor.CreateOrder(
					futures.OrderTypeLimit,
					futures.SideTypeSell,
					futures.TimeInForceTypeGTC,
					quantity,
					false,
					false,
					upPrice,
					upPrice,
					0,
					0)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Sell order %v on price %v status %v quantity %v",
					pairProcessor.GetPair(), order.GetOrderId(), upOrder.OrderID, upPrice, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upRecord := grid_types.NewRecord(upOrder.OrderID, upPrice, quantity, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upRecord)
				order.SetUpPrice(upPrice) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status == futures.OrderStatusTypeFilled {
					takerOrder = upOrder
				}
			} else {
				if pairProcessor.GetUpBound() == 0 || upPrice > pairProcessor.GetUpBound() {
					logrus.Debugf("Futures %s: UpBound %v isn't 0 and price %v > UpBound %v",
						pairProcessor.GetPair(), pairProcessor.GetUpBound(), upPrice, pairProcessor.GetUpBound())
				} else if depth_items.PriceType(utils.ConvStrToFloat64(risk.IsolatedMargin)) > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance())
				} else if locked > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pairProcessor.GetPair(), locked, pairProcessor.GetFreeBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель нижче
		downPrice, ok := grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		if ok && downPrice.GetOrderId() == 0 && downPrice.GetQuantity() <= 0 {
			// Створюємо ордер на купівлю
			downOrder, err := pairProcessor.CreateOrder(
				futures.OrderTypeLimit,
				futures.SideTypeBuy,
				futures.TimeInForceTypeGTC,
				quantity,
				false,
				false,
				order.GetDownPrice(),
				order.GetDownPrice(),
				0,
				0)
			if err != nil {
				printError()
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetQuantity(quantity)           // Записуємо кількість ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
				pairProcessor.GetPair(), order.GetOrderId(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status, quantity)
			if downOrder.Status == futures.OrderStatusTypeFilled {
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				pairProcessor,
				takerOrder.Side,
				grid,
				percentsToStopSettingNewOrder,
				order,
				quantity,
				locked,
				risk)
			if err != nil {
				printError()
				return err
			}
		}
	} else if side == futures.SideTypeBuy {
		// Якшо нижче немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetDownPrice() == 0 {
			// Створюємо ордер на купівлю
			downPrice := pairProcessor.RoundPrice(order.GetPrice() * (1 - pairProcessor.GetDeltaPrice()))
			if (pairProcessor.GetLowBound() == 0 || downPrice >= pairProcessor.GetLowBound()) &&
				float64(delta_percent(downPrice)) >= percentsToStopSettingNewOrder &&
				depth_items.PriceType(utils.ConvStrToFloat64(risk.IsolatedMargin)) <= pairProcessor.GetFreeBalance() &&
				locked <= pairProcessor.GetFreeBalance() {
				downOrder, err := pairProcessor.CreateOrder(
					futures.OrderTypeLimit,
					futures.SideTypeBuy,
					futures.TimeInForceTypeGTC,
					quantity,
					false,
					false,
					downPrice,
					downPrice,
					0,
					0)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
					pairProcessor.GetPair(), order.GetOrderId(), downOrder.OrderID, downPrice, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downRecord := grid_types.NewRecord(downOrder.OrderID, downPrice, quantity, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downRecord)
				order.SetDownPrice(downPrice) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status == futures.OrderStatusTypeFilled {
					takerRecord = downRecord
					takerOrder = downOrder
				}
			} else {
				if pairProcessor.GetLowBound() == 0 || downPrice < pairProcessor.GetLowBound() {
					logrus.Debugf("Futures %s: LowBound %v isn't 0 and price %v < LowBound %v",
						pairProcessor.GetPair(), pairProcessor.GetLowBound(), downPrice, pairProcessor.GetLowBound())
				} else if float64(delta_percent(downPrice)) < percentsToStopSettingNewOrder {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pairProcessor.GetPair(), risk.LiquidationPrice, delta_percent(downPrice), percentsToStopSettingNewOrder)
				} else if depth_items.PriceType(utils.ConvStrToFloat64(risk.IsolatedMargin)) > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pairProcessor.GetPair(), risk.IsolatedMargin, pairProcessor.GetFreeBalance())
				} else if locked > pairProcessor.GetFreeBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pairProcessor.GetPair(), locked, pairProcessor.GetFreeBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель вище
		upRecord, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upRecord.GetOrderId() == 0 && upRecord.GetQuantity() <= 0 {
			// Створюємо ордер на продаж
			upOrder, err := pairProcessor.CreateOrder(
				futures.OrderTypeLimit,
				futures.SideTypeSell,
				futures.TimeInForceTypeGTC,
				quantity,
				false,
				false,
				order.GetUpPrice(),
				order.GetUpPrice(),
				0,
				0)
			if err != nil {
				printError()
				return err
			}
			if upOrder.Status == futures.OrderStatusTypeFilled {
				takerRecord = upRecord
				takerOrder = upOrder
			}
			upRecord.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upRecord.SetQuantity(quantity)            // Записуємо кількість ордера в грід
			upRecord.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: From order %v Set Sell order %v on price %v status %v quantity %v",
				pairProcessor.GetPair(), order.GetOrderId(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				pairProcessor,
				takerOrder.Side,
				grid,
				percentsToStopSettingNewOrder,
				takerRecord,
				quantity,
				locked,
				risk)
			if err != nil {
				printError()
				return err
			}
		}
	}
	return
}
