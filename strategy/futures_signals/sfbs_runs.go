package futures_signals

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	// "github.com/fr0ster/go-trading-utils/binance/futures/account"
	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	types "github.com/fr0ster/go-trading-utils/types"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	deltaUp    = 0.0005
	deltaDown  = 0.0005
	degree     = 3
	limit      = 1000
	interval   = "1m"
	reloadTime = 500 * time.Millisecond
)

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

func RunFuturesHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	stopEvent <- os.Interrupt
	return fmt.Errorf("it should be implemented for futures")
}

func RunScalpingHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal) (err error) {
	pair.SetStrategy(pairs_types.GridStrategyType)
	return RunFuturesGridTrading(config, client, pair, stopEvent)
}

func RunFuturesTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}

	if config.GetConfigurations().GetReloadConfig() {
		go func() {
			for {
				<-time.After(reloadTime)
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
			}
		}()
	}

	stopEvent <- os.Interrupt
	return fmt.Errorf("it hadn't been implemented yet")
}

// Створення ордера для розміщення в грід
func createOrderInGrid(
	pairProcessor *PairProcessor,
	side futures.SideType,
	quantity,
	price float64) (order *futures.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		futures.OrderTypeLimit,     // orderType
		side,                       // sideType
		futures.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		false,                      // closePosition
		price,                      // price
		0,                          // stopPrice
		0)                          // callbackRate
	return
}

// Округлення ціни до StepSize знаків після коми
func getStepSizeExp(symbol *futures.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func getTickSizeExp(symbol *futures.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
}

func round(val float64, exp int) float64 {
	return utils.RoundToDecimalPlace(val, exp)
}

func IsOrdersOpened(grid *grid_types.Grid, pairProcessor *PairProcessor, pair *pairs_types.Pairs) (err error) {
	grid.Ascend(func(item btree.Item) bool {
		var orderOut *futures.Order
		record := item.(*grid_types.Record)
		if record.GetOrderId() != 0 {
			orderOut, err = pairProcessor.GetOrder(record.GetOrderId())
			if err != nil {
				return false
			}
			if orderOut == nil || orderOut.Status != futures.OrderStatusTypeNew {
				err = fmt.Errorf("futures %s: Order %v not found or status %v", pair.GetPair(), record.GetOrderId(), orderOut.Status)
			}
		}
		return true
	})
	return err
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams,
	symbol *futures.Symbol,
	side futures.SideType,
	grid *grid_types.Grid,
	order *grid_types.Record,
	quantity float64,
	exp int,
	locked float64,
	risk *futures.PositionRisk) (err error) {
	var (
		takerRecord *grid_types.Record
		takerOrder  *futures.CreateOrderResponse
	)
	delta_percent := func(currentPrice float64) float64 {
		return math.Abs((currentPrice - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
	}
	if side == futures.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			upPrice := round(order.GetPrice()*(1+pair.GetSellDelta()), exp)
			if (pair.GetUpBound() == 0 || upPrice <= pair.GetUpBound()) &&
				delta_percent(upPrice) >= config.GetConfigurations().GetPercentsToStopSettingNewOrder() &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pair.GetCurrentPositionBalance() &&
				locked <= pair.GetCurrentPositionBalance() {
				upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, upPrice)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
					pair.GetPair(), upOrder.OrderID, upPrice, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upRecord := grid_types.NewRecord(upOrder.OrderID, upPrice, quantity, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upRecord)
				order.SetUpPrice(upPrice) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status == futures.OrderStatusTypeFilled ||
					(config.GetConfigurations().GetMaintainPartiallyFilledOrders() && upOrder.Status == futures.OrderStatusTypePartiallyFilled) {
					takerRecord = upRecord
					takerOrder = upOrder
				}
			} else {
				if pair.GetUpBound() == 0 || upPrice > pair.GetUpBound() {
					logrus.Debugf("Futures %s: UpBound %v isn't 0 and price %v > UpBound %v",
						pair.GetPair(), pair.GetUpBound(), upPrice, pair.GetUpBound())
				} else if delta_percent(upPrice) < config.GetConfigurations().GetPercentsToStopSettingNewOrder() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent(upPrice), config.GetConfigurations().GetPercentsToStopSettingNewOrder())
				} else if utils.ConvStrToFloat64(risk.IsolatedMargin) > pair.GetCurrentPositionBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
				} else if locked > pair.GetCurrentPositionBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pair.GetPair(), locked, pair.GetCurrentPositionBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель нижче
		downPrice, ok := grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		if ok && downPrice.GetOrderId() == 0 {
			// Створюємо ордер на купівлю
			downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, order.GetDownPrice())
			if err != nil {
				printError()
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: Set Buy order %v on price %v status %v quantity %v",
				pair.GetPair(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status, quantity)
			if downOrder.Status == futures.OrderStatusTypeFilled ||
				(config.GetConfigurations().GetMaintainPartiallyFilledOrders() && downOrder.Status == futures.OrderStatusTypePartiallyFilled) {
				takerRecord = downPrice
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair, pairStreams,
				symbol,
				takerOrder.Side,
				grid,
				takerRecord,
				quantity,
				exp,
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
			downPrice := round(order.GetPrice()*(1-pair.GetBuyDelta()), exp)
			if (pair.GetLowBound() == 0 || downPrice >= pair.GetLowBound()) &&
				delta_percent(downPrice) >= config.GetConfigurations().GetPercentsToStopSettingNewOrder() &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pair.GetCurrentPositionBalance() &&
				locked <= pair.GetCurrentPositionBalance() {
				downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, downPrice)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: Set Buy order %v on price %v status %v quantity %v",
					pair.GetPair(), downOrder.OrderID, downPrice, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downRecord := grid_types.NewRecord(downOrder.OrderID, downPrice, quantity, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downRecord)
				order.SetDownPrice(downPrice) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status == futures.OrderStatusTypeFilled ||
					(config.GetConfigurations().GetMaintainPartiallyFilledOrders() && downOrder.Status == futures.OrderStatusTypePartiallyFilled) {
					takerRecord = downRecord
					takerOrder = downOrder
				}
			} else {
				if pair.GetLowBound() == 0 || downPrice < pair.GetLowBound() {
					logrus.Debugf("Futures %s: LowBound %v isn't 0 and price %v < LowBound %v",
						pair.GetPair(), pair.GetLowBound(), downPrice, pair.GetLowBound())
				} else if delta_percent(downPrice) < config.GetConfigurations().GetPercentsToStopSettingNewOrder() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent(downPrice), config.GetConfigurations().GetPercentsToStopSettingNewOrder())
				} else if utils.ConvStrToFloat64(risk.IsolatedMargin) > pair.GetCurrentPositionBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v > current position balance %v",
						pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
				} else if locked > pair.GetCurrentPositionBalance() {
					logrus.Debugf("Futures %s: Locked %v > current position balance %v",
						pair.GetPair(), locked, pair.GetCurrentPositionBalance())
				}
			}
		}
		// Знаходимо у гріді відповідний запис, та записи на шабель вище
		upRecord, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upRecord.GetOrderId() == 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, order.GetUpPrice())
			if err != nil {
				printError()
				return err
			}
			if upOrder.Status == futures.OrderStatusTypeFilled ||
				(config.GetConfigurations().GetMaintainPartiallyFilledOrders() && upOrder.Status == futures.OrderStatusTypePartiallyFilled) {
				takerRecord = upRecord
				takerOrder = upOrder
			}
			upRecord.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upRecord.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
				pair.GetPair(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair, pairStreams,
				symbol,
				takerOrder.Side,
				grid,
				takerRecord,
				quantity,
				exp,
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

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal) (err error) {
	var (
		quantity     float64
		locked       float64
		free         float64
		currentPrice float64
	)
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.GridStrategyType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		stopEvent <- os.Interrupt
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	// Створюємо стрім подій
	pairStreams, err := NewPairStreams(client, pair, false)
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return
	}
	// Створюємо обробник пари
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), false)
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}

	balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	pair.SetCurrentBalance(balance)
	config.Save()
	if pair.GetInitialBalance() == 0 {
		pair.SetInitialBalance(balance)
		config.Save()
	}
	if pair.GetMarginType() == "" {
		logrus.Debugf("Futures %s set MarginType %v from account into config", pair.GetPair(), pairProcessor.GetMarginType())
		pair.SetMarginType(pairProcessor.GetMarginType())
		config.Save()
	} else {
		if pair.GetMarginType() != pairProcessor.GetMarginType() {
			logrus.Debugf("Futures %s set MarginType %v from config into account", pair.GetPair(), pair.GetMarginType())
			pairProcessor.SetMarginType(pair.GetMarginType())
		}
	}
	if pair.GetLeverage() == 0 {
		logrus.Debugf("Futures %s set Leverage %v from account into config", pair.GetPair(), pairProcessor.GetLeverage())
		pair.SetLeverage(pairProcessor.GetLeverage())
		config.Save()
	} else {
		if pair.GetLeverage() != pairProcessor.GetLeverage() {
			logrus.Debugf("Futures %s set Leverage %v from config into account", pair.GetPair(), pair.GetLeverage())
			pairProcessor.SetLeverage(pair.GetLeverage())
		}
	}
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	symbol, err := func() (res *futures.Symbol, err error) {
		val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()})
		if val == nil {
			return nil, fmt.Errorf("futures %s: Symbol not found", pair.GetPair())
		}
		return val.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
	}()
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	tickSizeExp := getTickSizeExp(symbol)
	stepSizeExp := getStepSizeExp(symbol)
	// Отримання середньої ціни
	price := round(pair.GetMiddlePrice(), tickSizeExp)
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	if entryPrice := utils.ConvStrToFloat64(risk.EntryPrice); entryPrice != 0 {
		price = round(entryPrice, tickSizeExp)
	}
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = round(price, tickSizeExp)
	}
	setQuantity := func(symbol *futures.Symbol) (quantity float64) {
		quantity = round(pair.GetCurrentBalance()*pair.GetLimitOnPosition()*pair.GetLimitOnTransaction()*float64(pair.GetLeverage())/price, stepSizeExp)
		minNotional := utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
		if quantity*price < minNotional {
			logrus.Debugf("Futures %s: Quantity %v * price %v < minNotional %v", pair.GetPair(), quantity, price, minNotional)
			quantity = round(minNotional/price, stepSizeExp)
		}
		return
	}
	quantity = setQuantity(symbol)
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, 0, round(price*(1+pair.GetSellDelta()), tickSizeExp), round(price*(1-pair.GetBuyDelta()), tickSizeExp), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pair.GetPair(), price)

	err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Створюємо ордери на продаж
	sellOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, round(price*(1+pair.GetSellDelta()), tickSizeExp))
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, round(price*(1+pair.GetSellDelta()), tickSizeExp), quantity, 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), round(price*(1+pair.GetSellDelta()), tickSizeExp))
	// Створюємо ордер на купівлю
	buyOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, round(price*(1-pair.GetBuyDelta()), tickSizeExp))
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, round(price*(1-pair.GetSellDelta()), tickSizeExp), quantity, price, 0, types.SideTypeBuy))
	grid.Debug("Futures Grid", "", pair.GetPair())
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			printError()
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			grid.Lock()
			// Знаходимо у гріді на якому був виконаний ордер
			order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
			if !ok {
				if !(event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled) {
					return fmt.Errorf("uncorrected order ID: %v", event.OrderTradeUpdate.ID)
				} else {
					continue // Вважаємо ордер обробили раніше???
				}
			}
			orderId := order.GetOrderId()
			order.SetQuantity(order.GetQuantity() - utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledQty))
			if order.GetQuantity() == 0 {
				// if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled ||
				// 	(config.GetConfigurations().GetMaintainPartiallyFilledOrders() && event.OrderTradeUpdate.Status == futures.OrderStatusTypePartiallyFilled) {
				currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
				account, _ := futures_account.New(client, degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
				if asset := account.GetAssets().Get(&futures_account.Asset{Asset: pair.GetBaseSymbol()}); asset != nil {
					locked = utils.ConvStrToFloat64(asset.(*futures_account.Asset).WalletBalance) - utils.ConvStrToFloat64(asset.(*futures_account.Asset).AvailableBalance)
					free = utils.ConvStrToFloat64(asset.(*futures_account.Asset).AvailableBalance)
				}
				logrus.Debugf("Futures %s: Order %v on price %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					grid.Unlock()
					stopEvent <- os.Interrupt
					printError()
					return
				}
				// Балансування маржі як треба
				if config.GetConfigurations().GetBalancingOfMargin() &&
					utils.ConvStrToFloat64(risk.IsolatedMargin) < pair.GetCurrentPositionBalance() {
					logrus.Debugf("Futures %s: IsolatedMargin %v < current position balance %v",
						pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
					err = pairProcessor.SetPositionMargin(pair.GetCurrentPositionBalance()-utils.ConvStrToFloat64(risk.IsolatedMargin), 1)
					if err != nil {
						grid.Unlock()
						stopEvent <- os.Interrupt
						printError()
						return err
					}
				}
				// Обробка наближення ліквідаціі
				if config.GetConfigurations().GetObservePriceLiquidation() {
					delta_percent := func(price float64) float64 {
						return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
					}
					delta := delta_percent(currentPrice)
					if delta < config.GetConfigurations().GetPercentToDecreasePosition() {
						logrus.Debugf("Futures %s: Distance to liquidation %f%% less than %f%%",
							pair.GetPair(), delta*100, config.GetConfigurations().GetPercentToDecreasePosition()*100)
						if free > pair.GetCurrentPositionBalance() {
							err = pairProcessor.SetPositionMargin(pair.GetCurrentPositionBalance(), 1)
							if err != nil {
								grid.Unlock()
								stopEvent <- os.Interrupt
								printError()
								return err
							}
							risk, err = pairProcessor.GetPositionRisk()
							if err != nil {
								grid.Unlock()
								stopEvent <- os.Interrupt
								printError()
								return err
							}
							logrus.Debugf("Futures %s: Old Margin %v, Add Margin %v, New Margin %v",
								pair.GetPair(), pair.GetCurrentPositionBalance(), free-pair.GetCurrentPositionBalance(), risk.IsolatedMargin)
						} else {
							logrus.Debugf("Futures %s: Free %v <= current position balance %v",
								pair.GetPair(), free, pair.GetCurrentPositionBalance())
							if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
								_, err = pairProcessor.CreateOrder(
									futures.OrderTypeMarket,    // orderType
									futures.SideTypeBuy,        // sideType
									futures.TimeInForceTypeGTC, // timeInForce
									quantity,                   // quantity
									false,                      // closePosition
									price,                      // price
									0,                          // stopPrice
									0)                          // callbackRate
							} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
								_, err = pairProcessor.CreateOrder(
									futures.OrderTypeMarket,    // orderType
									futures.SideTypeSell,       // sideType
									futures.TimeInForceTypeGTC, // timeInForce
									quantity,                   // quantity
									false,                      // closePosition
									price,                      // price
									0,                          // stopPrice
									0)                          // callbackRate
							}
							if err != nil {
								grid.Unlock()
								stopEvent <- os.Interrupt
								printError()
								return err
							}
							risk, err = pairProcessor.GetPositionRisk()
							if err != nil {
								grid.Unlock()
								stopEvent <- os.Interrupt
								printError()
								return err
							}
						}
					}
				}
				err = processOrder(
					config,
					pairProcessor,
					pair,
					pairStreams,
					symbol,
					event.OrderTradeUpdate.Side,
					grid,
					order,
					quantity,
					tickSizeExp,
					locked,
					risk)
				if err != nil {
					stopEvent <- os.Interrupt
					pairProcessor.CancelAllOrders()
					printError()
					return err
				}
				grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
			} else if event.Event == futures.UserDataEventTypeAccountUpdate {
				logrus.Debugf("Futures %s: Account Update", pair.GetPair())
			} else if event.Event == futures.UserDataEventTypeMarginCall {
				logrus.Debugf("Futures %s: Margin Call", pair.GetPair())
			}
			grid.Unlock()
		}
	}
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan os.Signal,
	debug bool) (err error) {
	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		return RunFuturesHolding(config, client, degree, limit, pair, stopEvent, time.Second, debug)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		return RunScalpingHolding(config, client, pair, stopEvent)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		return RunFuturesTrading(config, client, degree, limit, pair, stopEvent, time.Second, debug)

		// Відпрацьовуємо Grid стратегію
	} else if pair.GetStrategy() == pairs_types.GridStrategyType {
		return RunFuturesGridTrading(config, client, pair, stopEvent)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		return fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
	}
}
