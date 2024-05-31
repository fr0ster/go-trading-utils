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

// Округлення ціни до TickSize знаків після коми
func getExp(symbol *futures.Symbol) int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
}
func roundPrice(val float64, exp int) float64 {
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
	risk *futures.PositionRisk,
	currentPrice float64,
	delta_percent float64) (err error) {
	var (
		takerPrice *grid_types.Record
		takerOrder *futures.CreateOrderResponse
	)
	if side == futures.SideTypeSell {
		// Якшо вище немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetUpPrice() == 0 {
			// Створюємо ордер на продаж
			if (pair.GetUpBound() == 0 || currentPrice <= pair.GetUpBound()) &&
				delta_percent >= config.GetConfigurations().GetPercentsToLiquidation() &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pair.GetCurrentPositionBalance() &&
				locked <= pair.GetCurrentPositionBalance() {
				upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, currentPrice)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: Set Sell order %v on price %v status %v quantity %v",
					pair.GetPair(), upOrder.OrderID, currentPrice, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upPrice := grid_types.NewRecord(upOrder.OrderID, currentPrice, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upPrice)
				order.SetUpPrice(currentPrice) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status != futures.OrderStatusTypeNew {
					takerPrice = upPrice
					takerOrder = upOrder
				}
			} else {
				if pair.GetUpBound() == 0 || currentPrice > pair.GetUpBound() {
					logrus.Debugf("Futures %s: UpBound %v isn't 0 and price %v > UpBound %v",
						pair.GetPair(), pair.GetUpBound(), currentPrice, pair.GetUpBound())
				} else if delta_percent < config.GetConfigurations().GetPercentsToLiquidation() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent, config.GetConfigurations().GetPercentsToLiquidation())
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
			if downOrder.Status != futures.OrderStatusTypeNew {
				takerPrice = downPrice
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
				takerPrice,
				quantity,
				exp,
				locked,
				risk,
				currentPrice,
				delta_percent)
			if err != nil {
				printError()
				return err
			}
		}
	} else if side == futures.SideTypeBuy {
		// Якшо нижче немае запису про створений ордер, то створюємо його і робимо запис в грід
		if order.GetDownPrice() == 0 {
			// Створюємо ордер на купівлю
			if (pair.GetLowBound() == 0 || currentPrice >= pair.GetLowBound()) &&
				delta_percent >= config.GetConfigurations().GetPercentsToLiquidation() &&
				utils.ConvStrToFloat64(risk.IsolatedMargin) <= pair.GetCurrentPositionBalance() &&
				locked <= pair.GetCurrentPositionBalance() {
				downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, currentPrice)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: Set Buy order %v on price %v status %v quantity %v",
					pair.GetPair(), downOrder.OrderID, currentPrice, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downPrice := grid_types.NewRecord(downOrder.OrderID, currentPrice, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downPrice)
				order.SetDownPrice(currentPrice) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status != futures.OrderStatusTypeNew {
					takerPrice = downPrice
					takerOrder = downOrder
				}
			} else {
				if pair.GetLowBound() == 0 || currentPrice < pair.GetLowBound() {
					logrus.Debugf("Futures %s: LowBound %v isn't 0 and price %v < LowBound %v",
						pair.GetPair(), pair.GetLowBound(), currentPrice, pair.GetLowBound())
				} else if delta_percent < config.GetConfigurations().GetPercentsToLiquidation() {
					logrus.Debugf("Futures %s: Liquidation price %v, distance %v less than %v",
						pair.GetPair(), risk.LiquidationPrice, delta_percent, config.GetConfigurations().GetPercentsToLiquidation())
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
		upPrice, ok := grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		if ok && upPrice.GetOrderId() == 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, order.GetUpPrice())
			if err != nil {
				printError()
				return err
			}
			if upOrder.Status != futures.OrderStatusTypeNew {
				takerPrice = upPrice
				takerOrder = upOrder
			}
			upPrice.SetOrderId(upOrder.OrderID)      // Записуємо номер ордера в грід
			upPrice.SetOrderSide(types.SideTypeSell) // Записуємо сторону ордера в грід
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
				takerPrice,
				quantity,
				exp,
				locked,
				risk,
				currentPrice,
				delta_percent)
			if err != nil {
				printError()
				return err
			}
		}
	}
	return
}

func observePriceLiquidation(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams,
	grid *grid_types.Grid,
	delta_percent float64) (err error) {
	if config.GetConfigurations().GetObservePriceLiquidation() {
		risk, err := pairStreams.GetPositionRisk()
		if err != nil {
			printError()
			return err
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
			// delta_percent := pairStreams.GetLiquidationDistance(price)
			if delta_percent <= config.GetConfigurations().GetPercentsToLiquidation() {
				logrus.Debugf("Futures %s: Liquidation price %v, delta %v!!!!!", pair.GetPair(), risk.LiquidationPrice, delta_percent)
				// Перевіряємо чи є зайві відкриті ордери
				// У випадку коли позиція від'ємна та є відкриті ордери на продаж, то відміняємо їх
				// ...або позиція позитивна та є відкриті ордери на купівлю, то відміняємо їх
				if (grid.GetCountBuyOrders() > 0 && utils.ConvStrToFloat64(risk.PositionAmt) > 0) ||
					(grid.GetCountSellOrders() > 0 && utils.ConvStrToFloat64(risk.PositionAmt) < 0) {
					grid.Lock()
					grid.Ascend(func(item btree.Item) bool {
						record := item.(*grid_types.Record)
						if record.GetOrderId() != 0 {
							if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
								if record.GetOrderSide() == types.SideTypeBuy {
									_, _ = pairProcessor.CancelOrder(record.GetOrderId())
								}
							} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
								if record.GetOrderSide() == types.SideTypeSell {
									_, _ = pairProcessor.CancelOrder(record.GetOrderId())
								}
							}
						}
						grid.Debug(pair.GetPair(), "", "Futures Liquidation")
						return true
					})
					if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
						grid.CancelBuyOrder()
					} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
						grid.CancelSellOrder()
					}
					grid.Unlock()
				} else { // Якщо немає відкритих ордерів, то перевіряємо чи є вільні кошти для збільшення маржі
					free, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
					if err != nil {
						return err
					}
					if free >= pair.GetCurrentPositionBalance() { // Як є вільні кошти, то збільшуємо маржу
						logrus.Debugf("Futures %s: Free asset %v >= current balance %v", pair.GetPair(), free, pair.GetCurrentPositionBalance())
						// Устанавлюемо Margin
						delta := pair.GetCurrentPositionBalance() - utils.ConvStrToFloat64(risk.IsolatedMargin)
						if delta != 0 {
							err = pairProcessor.SetPositionMargin(delta, 1)
							if err != nil {
								logrus.Errorf("Futures %s: Set position margin error %v in observePriceLiquidation", pair.GetPair(), err)
								return err
							}
						}
					} else { // Як немає вільних коштів, то зменшуємо позицію
						logrus.Debugf("Futures %s: Free asset %v < current balance %v", pair.GetPair(), free, pair.GetCurrentPositionBalance())
						positionAmtDec := utils.ConvStrToFloat64(risk.PositionAmt) * delta_percent
						if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
							logrus.Debugf("Futures %s: Liquidation price %v, delta %v, position %v, new position %v",
								pair.GetPair(), risk.LiquidationPrice, delta_percent, risk.PositionAmt, positionAmtDec)
							_, err = pairProcessor.CreateOrder(
								futures.OrderTypeMarket,    // orderType
								futures.SideTypeSell,       // sideType
								futures.TimeInForceTypeGTC, // timeInForce
								positionAmtDec,             // quantity
								false,                      // closePosition
								0,                          // price
								0,                          // stopPrice
								0)                          // callbackRate
						} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
							logrus.Debugf("Futures %s: Liquidation price %v, delta %v, position %v, new position %v",
								pair.GetPair(), risk.LiquidationPrice, delta_percent, risk.PositionAmt, positionAmtDec)
							_, err = pairProcessor.CreateOrder(
								futures.OrderTypeMarket,    // orderType
								futures.SideTypeBuy,        // sideType
								futures.TimeInForceTypeGTC, // timeInForce
								positionAmtDec,             // quantity
								false,                      // closePosition
								0,                          // price
								0,                          // stopPrice
								0)                          // callbackRate
						}
						printError()
						return err
					}
				}
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
		free          float64
		locked        float64
		currentPrice  float64
		delta_percent float64
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
	exp := getExp(symbol)
	// Отримання середньої ціни
	price := roundPrice(pair.GetMiddlePrice(), exp)
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	if entryPrice := utils.ConvStrToFloat64(risk.EntryPrice); entryPrice != 0 {
		price = roundPrice(entryPrice, exp)
	}
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = roundPrice(price, exp)
	}
	setQuantity := func(symbol *futures.Symbol) (quantity float64) {
		quantity = pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() * float64(pair.GetLeverage()) / price
		minNotional := utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
		if quantity*price < minNotional {
			logrus.Debugf("Futures %s: Quantity %v * price %v < minNotional %v", pair.GetPair(), quantity, price, minNotional)
			quantity = minNotional / price
		}
		return
	}
	quantity := setQuantity(symbol)
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, roundPrice(price*(1+pair.GetSellDelta()), exp), roundPrice(price*(1-pair.GetBuyDelta()), exp), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pair.GetPair(), price)

	err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Створюємо ордери на продаж
	sellOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, roundPrice(price*(1+pair.GetSellDelta()), exp))
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, roundPrice(price*(1+pair.GetSellDelta()), exp), 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), roundPrice(price*(1+pair.GetSellDelta()), exp))
	// Створюємо ордер на купівлю
	buyOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, roundPrice(price*(1-pair.GetBuyDelta()), exp))
	if err != nil {
		stopEvent <- os.Interrupt
		printError()
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, roundPrice(price*(1-pair.GetSellDelta()), exp), price, 0, types.SideTypeBuy))
	// Запускаємо спостереження за залоченими коштами та оновлення конфігурації
	go func() {
		for {
			<-time.After(time.Duration(config.GetConfigurations().GetObserverTimeOut()) * time.Millisecond)
			// free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
			// locked, _ = pairStreams.GetAccount().GetLockedAsset(pair.GetBaseSymbol())
			// risk, err = pairProcessor.GetPositionRisk()
			// if err != nil {
			// 	stopEvent <- os.Interrupt
			// 	printError()
			// 	return
			// }
			// Спостереження за ліквідацією при потребі
			if currentPrice != 0 && risk != nil {
				delta_percent = math.Abs((currentPrice - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
				err = observePriceLiquidation(config, pairProcessor, pair, pairStreams, grid, delta_percent)
				if err != nil {
					stopEvent <- os.Interrupt
					printError()
					return
				}
			}
			if config.GetConfigurations().GetReloadConfig() {
				config.Load()
				pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
				balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
				if err != nil {
					printError()
					return
				}
				pair.SetCurrentBalance(balance)
				config.Save()
				quantity = setQuantity(symbol)
			}
		}
	}()
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
			if event.Event == futures.UserDataEventTypeOrderTradeUpdate {
				grid.Lock()
				currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
				// free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
				// locked, _ = pairStreams.GetAccount().GetLockedAsset(pair.GetBaseSymbol())
				// risk, err = pairProcessor.GetPositionRisk()
				// if err != nil {
				// 	stopEvent <- os.Interrupt
				// 	printError()
				// 	return
				// }
				account, _ := futures_account.New(client, degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
				if asset := account.GetAssets().Get(&futures_account.Asset{Asset: pair.GetBaseSymbol()}); asset != nil {
					free = utils.ConvStrToFloat64(asset.(*futures_account.Asset).WalletBalance)
					locked = utils.ConvStrToFloat64(asset.(*futures_account.Asset).WalletBalance) - utils.ConvStrToFloat64(asset.(*futures_account.Asset).AvailableBalance)
				}
				logrus.Debugf("Futures %s: Order %v on price %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				if risk == nil || utils.ConvStrToFloat64(risk.PositionAmt) == 0 {
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						printError()
						return
					}
					delta_percent = math.Abs((currentPrice - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
				}
				if utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
					if delta_percent <= config.GetConfigurations().GetPercentsToLiquidation() {
						if utils.ConvStrToFloat64(risk.IsolatedMargin) < pair.GetCurrentPositionBalance() {
							delta := pair.GetCurrentPositionBalance() - utils.ConvStrToFloat64(risk.IsolatedMargin)
							if delta != 0 && free > delta {
								err = pairProcessor.SetPositionMargin(delta, 1)
								if err != nil {
									logrus.Errorf("Futures %s: New Margin %v, Old Margin %v, IsAutoAddMargin %v, Free %v error %v in event maintainer",
										pair.GetPair(), delta, risk.IsolatedMargin, risk.IsAutoAddMargin, free, err)
									printError()
									return
								}
								logrus.Debugf("Futures %s: Margin was %v, add Margin %v",
									pair.GetPair(), utils.ConvStrToFloat64(risk.IsolatedMargin), delta)
							} else {
								logrus.Debugf("Futures %s: Free asset %v < delta %v", pair.GetPair(), free, delta)
							}
						}
					}
				}
				// Знаходимо у гріді на якому був виконаний ордер
				order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
				if !ok {
					return fmt.Errorf("uncorrected order ID: %v", event.OrderTradeUpdate.ID)
				}
				orderId := order.GetOrderId()
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
					exp,
					locked,
					risk,
					currentPrice,
					delta_percent)
				if err != nil {
					pairProcessor.CancelAllOrders()
					printError()
					return err
				}
				grid.Unlock()
				grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
			} else if event.Event == futures.UserDataEventTypeAccountUpdate {
				logrus.Debugf("Futures %s: Account Update", pair.GetPair())
			} else if event.Event == futures.UserDataEventTypeMarginCall {
				logrus.Debugf("Futures %s: Margin Call", pair.GetPair())
			}
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
