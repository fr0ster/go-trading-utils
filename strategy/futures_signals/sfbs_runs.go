package futures_signals

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	types "github.com/fr0ster/go-trading-utils/types"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

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
	quit chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	if pair.GetAccountType() != pairs_types.USDTFutureType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	return fmt.Errorf("it should be implemented for futures")
}

func RunScalpingHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	pair.SetStrategy(pairs_types.GridStrategyType)
	return RunFuturesGridTrading(config, client, pair, quit, wg)
}

func RunFuturesTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	quit chan struct{},
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
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
				select {
				case <-quit:
					logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
					return
				case <-time.After(reloadTime):
					config.Load()
					pair = config.GetConfigurations().GetPair(pair.GetAccountType(), pair.GetStrategy(), pair.GetStage(), pair.GetPair())
				}
			}
		}()
	}

	return fmt.Errorf("it hadn't been implemented yet")
}

// Створення ордера для розміщення в грід
func createOrder(
	pairProcessor *PairProcessor,
	side futures.SideType,
	orderType futures.OrderType,
	quantity,
	price float64,
	callbackRate float64,
	closePosition bool) (order *futures.CreateOrderResponse, err error) {
	order, err = pairProcessor.CreateOrder(
		orderType,                  // orderType
		side,                       // sideType
		futures.TimeInForceTypeGTC, // timeInForce
		quantity,                   // quantity
		closePosition,              // closePosition
		price,                      // price
		price,                      // stopPrice
		callbackRate)               // callbackRate
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
	// pairStreams *PairStreams,
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
				upOrder, err := createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeLimit, quantity, upPrice, 0, false)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Sell order %v on price %v status %v quantity %v",
					pair.GetPair(), order.GetOrderId(), upOrder.OrderID, upPrice, upOrder.Status, quantity)
				// Записуємо ордер в грід
				upRecord := grid_types.NewRecord(upOrder.OrderID, upPrice, quantity, 0, order.GetPrice(), types.OrderSide(futures.SideTypeSell))
				grid.Set(upRecord)
				order.SetUpPrice(upPrice) // Ставимо посилання на верхній запис в гріді
				if upOrder.Status == futures.OrderStatusTypeFilled {
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
		if ok && downPrice.GetOrderId() == 0 && downPrice.GetQuantity() <= 0 {
			// Створюємо ордер на купівлю
			downOrder, err := createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeLimit, quantity, order.GetDownPrice(), 0, false)
			if err != nil {
				printError()
				return err
			}
			downPrice.SetOrderId(downOrder.OrderID)   // Записуємо номер ордера в грід
			downPrice.SetQuantity(quantity)           // Записуємо кількість ордера в грід
			downPrice.SetOrderSide(types.SideTypeBuy) // Записуємо сторону ордера в грід
			logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
				pair.GetPair(), order.GetOrderId(), downOrder.OrderID, order.GetDownPrice(), downOrder.Status, quantity)
			if downOrder.Status == futures.OrderStatusTypeFilled {
				takerRecord = downPrice
				takerOrder = downOrder
			}
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair,
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
				downOrder, err := createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeLimit, quantity, downPrice, 0, false)
				if err != nil {
					printError()
					return err
				}
				logrus.Debugf("Futures %s: From order %v Set Buy order %v on price %v status %v quantity %v",
					pair.GetPair(), order.GetOrderId(), downOrder.OrderID, downPrice, downOrder.Status, quantity)
				// Записуємо ордер в грід
				downRecord := grid_types.NewRecord(downOrder.OrderID, downPrice, quantity, order.GetPrice(), 0, types.OrderSide(futures.SideTypeBuy))
				grid.Set(downRecord)
				order.SetDownPrice(downPrice) // Ставимо посилання на нижній запис в гріді
				if downOrder.Status == futures.OrderStatusTypeFilled {
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
		if ok && upRecord.GetOrderId() == 0 && upRecord.GetQuantity() <= 0 {
			// Створюємо ордер на продаж
			upOrder, err := createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeLimit, quantity, order.GetUpPrice(), 0, false)
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
				pair.GetPair(), order.GetOrderId(), upOrder.OrderID, order.GetUpPrice(), upOrder.Status, quantity)
		}
		order.SetOrderId(0)                    // Помічаємо ордер як виконаний
		order.SetQuantity(0)                   // Помічаємо ордер як виконаний
		order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
		if takerOrder != nil {
			err = processOrder(
				config,
				pairProcessor,
				pair,
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

func checkRun(
	pair *pairs_types.Pairs,
	accountType pairs_types.AccountType,
	strategyType pairs_types.StrategyType) error {
	if pair.GetAccountType() != accountType {
		printError()
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != strategyType {
		printError()
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		printError()
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	return nil
}

func loadConfig(pair *pairs_types.Pairs, config *config_types.ConfigFile, pairProcessor *PairProcessor) (err error) {
	baseValue, err := pairProcessor.GetFreeBalance()
	pair.SetCurrentBalance(baseValue)
	config.Save()
	if pair.GetInitialBalance() == 0 {
		pair.SetInitialBalance(baseValue)
		config.Save()
	}
	return
}

func initRun(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{}) (pairProcessor *PairProcessor, err error) {
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(config, client, pair, quit, false)
	if err != nil {
		printError()
		return
	}

	balance, err := pairProcessor.GetFreeBalance()
	if err != nil {
		printError()
		return
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
	return
}

func updateConfig(config *config_types.ConfigFile, pair *pairs_types.Pairs) {
	if config.GetConfigurations().GetReloadConfig() {
		temp := config_types.NewConfigFile(config.GetFileName())
		temp.Load()
		t_pair := config.GetConfigurations().GetPair(
			pair.GetAccountType(),
			pair.GetStrategy(),
			pair.GetStage(),
			pair.GetPair())

		pair.SetLimitOnPosition(t_pair.GetLimitOnPosition())
		pair.SetLimitOnTransaction(t_pair.GetLimitOnTransaction())
		pair.SetSellDelta(t_pair.GetSellDelta())
		pair.SetBuyDelta(t_pair.GetBuyDelta())
		pair.SetUpBound(t_pair.GetUpBound())
		pair.SetLowBound(t_pair.GetLowBound())

		config.GetConfigurations().SetLogLevel(temp.GetConfigurations().GetLogLevel())
		config.GetConfigurations().SetReloadConfig(temp.GetConfigurations().GetReloadConfig())
		config.GetConfigurations().SetObservePriceLiquidation(temp.GetConfigurations().GetObservePriceLiquidation())
		config.GetConfigurations().SetObservePositionLoss(temp.GetConfigurations().GetObservePositionLoss())
		config.GetConfigurations().SetClosePositionOnRestart(temp.GetConfigurations().GetClosePositionOnRestart())
		config.GetConfigurations().SetBalancingOfMargin(temp.GetConfigurations().GetBalancingOfMargin())
		config.GetConfigurations().SetPercentsToStopSettingNewOrder(temp.GetConfigurations().GetPercentsToStopSettingNewOrder())
		config.GetConfigurations().SetPercentToDecreasePosition(temp.GetConfigurations().GetPercentToDecreasePosition())
		config.GetConfigurations().SetObserverTimeOutMillisecond(temp.GetConfigurations().GetObserverTimeOutMillisecond())
		config.GetConfigurations().SetUsingBreakEvenPrice(temp.GetConfigurations().GetUsingBreakEvenPrice())

		config.Save()
	}
}

func getSymbol(
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor) (res *futures.Symbol, err error) {
	val := pairProcessor.GetSymbol()
	if val == nil {
		printError()
		return nil, fmt.Errorf("futures %s: Symbol not found", pair.GetPair())
	}
	return val.GetFuturesSymbol()
}

func initVars(
	isDynamicDelta bool,
	client *futures.Client,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor) (
	symbol *futures.Symbol,
	price float64,
	priceUp,
	priceDown,
	quantity float64,
	minNotional float64,
	tickSizeExp,
	stepSizeExp int,
	err error) {
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		err = fmt.Errorf("futures %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
		printError()
		return
	}
	symbol, err = getSymbol(pair, pairProcessor)
	if err != nil {
		printError()
		return
	}
	tickSizeExp = getTickSizeExp(symbol)
	stepSizeExp = getStepSizeExp(symbol)
	// Отримання середньої ціни
	price = round(pair.GetMiddlePrice(), tickSizeExp)
	if price <= 0 {
		price, _ = GetCurrentPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = round(price, tickSizeExp)
	}
	// Отримання ціни по відкритій позиції
	risk, err := pairProcessor.GetPositionRisk()
	if err != nil {
		printError()
		return
	}
	if isDynamicDelta && risk != nil && utils.ConvStrToFloat64(risk.BreakEvenPrice) != 0 {
		priceUp = round(math.Max(utils.ConvStrToFloat64(risk.BreakEvenPrice), price)*(1+pair.GetSellDelta()), tickSizeExp)
		priceDown = round(math.Min(utils.ConvStrToFloat64(risk.BreakEvenPrice), price)*(1-pair.GetBuyDelta()), tickSizeExp)
	} else {
		priceUp = round(price*(1+pair.GetSellDelta()), tickSizeExp)
		priceDown = round(price*(1-pair.GetBuyDelta()), tickSizeExp)
	}
	setQuantity := func(symbol *futures.Symbol) (quantity float64) {
		quantity = round(pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction()*float64(pair.GetLeverage())/price, stepSizeExp)
		minNotional = utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
		if quantity*price < minNotional {
			logrus.Debugf("Futures %s: Quantity %v * price %v < minNotional %v", pair.GetPair(), quantity, price, minNotional)
			quantity = round(minNotional/price, stepSizeExp)
		}
		return
	}
	quantity = setQuantity(symbol)
	logrus.Debugf("Futures %s: Initial price %v, Quantity %v, MinNotional %v, TickSizeExp %v, StepSizeExp %v",
		pair.GetPair(), price, quantity, minNotional, tickSizeExp, stepSizeExp)
	return
}

func openPosition(
	pair *pairs_types.Pairs,
	quantity float64,
	priceUp float64,
	priceDown float64,
	pairProcessor *PairProcessor) (sellOrder, buyOrder *futures.CreateOrderResponse, err error) {
	err = pairProcessor.CancelAllOrders()
	if err != nil {
		printError()
		return
	}
	// Створюємо ордери на продаж
	sellOrder, err = createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeLimit, quantity, priceUp, 0, false)
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set Sell order on price %v with quantity %v status %v", pair.GetPair(), priceUp, quantity, sellOrder.Status)
	// Створюємо ордери на купівлю
	buyOrder, err = createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeLimit, quantity, priceDown, 0, false)
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set Buy order on price %v with quantity %v status %v", pair.GetPair(), priceDown, quantity, buyOrder.Status)
	return
}

func getCurrentPrice(
	client *futures.Client,
	pair *pairs_types.Pairs,
	tickSizeExp int) (currentPrice float64) {
	val, _ := GetCurrentPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
	currentPrice = round(val, tickSizeExp)
	return
}

func marginBalancing(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor,
	free float64,
	tickStepSize int) (freeOut float64, err error) {
	// Балансування маржі як треба
	if config.GetConfigurations().GetBalancingOfMargin() && utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
		delta := round(pair.GetCurrentPositionBalance(), tickStepSize) - round(utils.ConvStrToFloat64(risk.IsolatedMargin), tickStepSize)
		if delta != 0 {
			if delta > 0 && delta < free {
				err = pairProcessor.SetPositionMargin(delta, 1)
				logrus.Debugf("Futures %s: IsolatedMargin %v < current position balance %v and we have enough free %v",
					pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance(), free)
			}
		}
		freeOut, _ = pairProcessor.GetFreeBalance()
	} else {
		freeOut = free
	}
	return
}

func liquidationObservation(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor,
	currentPrice float64,
	free float64,
	price float64,
	quantity float64) (err error) {
	// Обробка наближення ліквідаціі
	if config.GetConfigurations().GetObservePriceLiquidation() &&
		utils.ConvStrToFloat64(risk.PositionAmt) != 0 {
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
					printError()
					return err
				}
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
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
					printError()
					return err
				}
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					return err
				}
			}
		}
	}
	return
}

func reOpenPosition(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor,
	quantity float64,
	price float64,
	tickSizeExp int) (err error) {
	var (
		side futures.SideType
	)
	// Обробка втрат по позиції
	if config.GetConfigurations().GetObservePositionLoss() {
		openOrders, _ := pairProcessor.GetOpenOrders()
		// Якщо є тільки один ордер, це означає що ціна може піти занадто далеко
		// шоб чекати на повернення і краще рестартувати з нового рівня
		if len(openOrders) == 1 {
			sellDeltaPercent := 0.0
			buyDeltaPercent := 0.0
			for _, order := range openOrders {
				if order.Side == futures.SideTypeSell {
					sellDeltaPercent = math.Abs(utils.ConvStrToFloat64(order.Price)-price) / price
				} else if order.Side == futures.SideTypeBuy {
					buyDeltaPercent = math.Abs(utils.ConvStrToFloat64(order.Price)-price) / price
				}
			}
			// Позиція від'ємна
			// if utils.ConvStrToFloat64(risk.UnRealizedProfit) < 0 &&
			// 	// Позиція більша за встановлений ліміт, тобто потенційна втрата більша за встановлений ліміт
			// 	math.Abs(utils.ConvStrToFloat64(risk.UnRealizedProfit)) > pair.GetCurrentPositionBalance()*(1+pair.GetUnRealizedProfitLowBound()) {
			if sellDeltaPercent > config.GetConfigurations().GetSellDeltaLoss() || buyDeltaPercent > config.GetConfigurations().GetBuyDeltaLoss() {
				// Скасовуємо всі ордери
				pairProcessor.CancelAllOrders()
				if config.GetConfigurations().GetClosePositionOnRestart() {
					// Закриваємо позицію
					if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
						side = futures.SideTypeSell
					} else if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
						side = futures.SideTypeBuy
					}
					_, err = pairProcessor.ClosePosition(side, price, tickSizeExp)
					if err != nil {
						printError()
						return err
					}
				}
				// Створюємо початкові ордери на продаж та купівлю з новими цінами
				priceUp := round(price*(1+pair.GetSellDelta()), tickSizeExp)
				priceDown := round(price*(1-pair.GetBuyDelta()), tickSizeExp)
				_, _, err = openPosition(pair, quantity, priceUp, priceDown, pairProcessor)
				if err != nil {
					printError()
					return err
				}
			}
		}
	}
	return err
}

func initGrid(
	pair *pairs_types.Pairs,
	price float64,
	quantity float64,
	tickSizeExp int,
	sellOrder, buyOrder *futures.CreateOrderResponse) (grid *grid_types.Grid, err error) {
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pair.GetPair())
	grid = grid_types.New()
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, 0, round(price*(1+pair.GetSellDelta()), tickSizeExp), round(price*(1-pair.GetBuyDelta()), tickSizeExp), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pair.GetPair(), price)
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, round(price*(1+pair.GetSellDelta()), tickSizeExp), quantity, 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), round(price*(1+pair.GetSellDelta()), tickSizeExp))
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, round(price*(1-pair.GetSellDelta()), tickSizeExp), quantity, price, 0, types.SideTypeBuy))
	grid.Debug("Futures Grid", "", pair.GetPair())
	return
}

func getCallBack_v1(
	config *config_types.ConfigFile,
	symbol *futures.Symbol,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	tickSizeExp int,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		free         float64
		currentPrice float64
		risk         *futures.PositionRisk
		err          error
	)
	return func(event *futures.WsUserDataEvent) {
		if grid == nil {
			return
		}
		if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				grid.Lock()
				updateConfig(config, pair)
				logrus.Debugf("Futures %s: Order %v on price %v with quantity %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
				// Знаходимо у гріді на якому був виконаний ордер
				order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
				if ok {
					orderId := order.GetOrderId()
					locked, _ = pairProcessor.GetLockedBalance()
					free, _ = pairProcessor.GetFreeBalance()
					pair.SetCurrentBalance(free)
					config.Save()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						grid.Unlock()
						printError()
						close(quit)
						return
					}
					// Балансування маржі як треба
					free, _ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
					// Обробка наближення ліквідаціі
					err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, currentPrice, quantity)
					if err != nil {
						grid.Unlock()
					}
					err = processOrder(
						config,
						pairProcessor,
						pair,
						symbol,
						event.OrderTradeUpdate.Side,
						grid,
						order,
						quantity,
						tickSizeExp,
						locked,
						risk)
					if err != nil {
						grid.Unlock()
						pairProcessor.CancelAllOrders()
						printError()
						close(quit)
						return
					}
					grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
					grid.Unlock()
				}
			}
		}
	}
}

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity    float64
		minNotional float64
		grid        *grid_types.Grid
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyType)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairProcessor, err := initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		return err
	}
	symbol, initPrice, initPriceUp, initPriceDown, quantity, minNotional, tickSizeExp, _, err := initVars(
		config.GetConfigurations().GetDynamicDelta(),
		client,
		pair,
		pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pair.GetCurrentPositionBalance(), pair.GetLimitOnTransaction())
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	_, err = streamStart(
		client,
		getCallBack_v1(
			config,
			symbol,
			pair,
			pairProcessor,
			grid,
			tickSizeExp,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(pair, quantity, initPriceUp, initPriceDown, pairProcessor)
	if err != nil {
		printError()
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pair, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
	if err != nil {
		printError()
		return err
	}
	grid.Debug("Futures Grid", "", pair.GetPair())
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		printError()
		return err
	}
	pairProcessor.CancelAllOrders()
	return nil
}

func getCallBack_v2(
	config *config_types.ConfigFile,
	symbol *futures.Symbol,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	grid *grid_types.Grid,
	tickSizeExp int,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	var (
		quantity     float64
		locked       float64
		free         float64
		currentPrice float64
		risk         *futures.PositionRisk
		err          error
	)
	return func(event *futures.WsUserDataEvent) {
		if grid == nil {
			return
		}
		if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			grid.Lock()
			updateConfig(config, pair)
			// Знаходимо у гріді на якому був виконаний ордер
			currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
			order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
			if !ok {
				printError()
				logrus.Errorf("we didn't work with order on price level %v before: %v", currentPrice, event.OrderTradeUpdate.ID)
				return
			}
			orderId := order.GetOrderId()
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				logrus.Debugf("Futures %s: Order %v on price %v with quantity %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
				locked, _ = pairProcessor.GetLockedBalance()
				free, _ = pairProcessor.GetFreeBalance()
				pair.SetCurrentBalance(free)
				config.Save()
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				// Балансування маржі як треба
				free, _ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
				// Обробка наближення ліквідаціі
				err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, currentPrice, quantity)
				if err != nil {
					grid.Unlock()
					printError()
					close(quit)
					return
				}
				err = processOrder(
					config,
					pairProcessor,
					pair,
					symbol,
					event.OrderTradeUpdate.Side,
					grid,
					order,
					quantity,
					tickSizeExp,
					locked,
					risk)
				if err != nil {
					grid.Unlock()
					pairProcessor.CancelAllOrders()
					printError()
					close(quit)
					return
				}
				grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
			}
			grid.Unlock()
		}
	}
}

func RunFuturesGridTradingV2(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity float64
		grid     *grid_types.Grid
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV2)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairProcessor, err := initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		return err
	}
	symbol, initPrice, initPriceUp, initPriceDown, quantity, minNotional, tickSizeExp, _, err := initVars(
		config.GetConfigurations().GetDynamicDelta(),
		client,
		pair,
		pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pair.GetCurrentPositionBalance(), pair.GetLimitOnTransaction())
	}
	maintainedOrders := btree.New(2)
	_, err = streamStart(
		client,
		getCallBack_v2(
			config,
			symbol,
			pair,
			pairProcessor,
			grid,
			tickSizeExp,
			quit,
			maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	sellOrder, buyOrder, err := openPosition(pair, quantity, initPriceUp, initPriceDown, pairProcessor)
	if err != nil {
		return err
	}
	// Ініціалізація гріду
	grid, err = initGrid(pair, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
	grid.Debug("Futures Grid", "", pair.GetPair())
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		printError()
		return err
	}
	pairProcessor.CancelAllOrders()
	return nil
}

func GetPricePair(
	isDynamicDelta bool,
	pair *pairs_types.Pairs,
	lastFilledPrice float64,
	lastExecutedSide futures.SideType,
	risk *futures.PositionRisk,
	currentPrice float64,
	tickSizeExp int) (upPrice, downPrice float64) {
	if isDynamicDelta {
		breakEvenPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
		// Визначаємо ціну для нових ордерів коли позиція від'ємна
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			if lastExecutedSide == futures.SideTypeSell {
				delta := currentPrice - lastFilledPrice
				upPrice = round(math.Max(lastFilledPrice, currentPrice)*(1+pair.GetSellDelta())+delta, tickSizeExp)
				downPrice = round(math.Min(breakEvenPrice, currentPrice)*(1-pair.GetBuyDelta()), tickSizeExp)
			} else if lastExecutedSide == futures.SideTypeBuy {
				upPrice = round(math.Max(lastFilledPrice, currentPrice)*(1+pair.GetSellDelta()), tickSizeExp)
				downPrice = round(math.Min(breakEvenPrice, currentPrice)*(1-pair.GetBuyDelta()), tickSizeExp)
			}
			// Визначаємо ціну для нових ордерів коли позиція позитивна
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			if lastExecutedSide == futures.SideTypeSell {
				upPrice = round(math.Max(breakEvenPrice, currentPrice)*(1+pair.GetSellDelta()), tickSizeExp)
				downPrice = round(math.Min(lastFilledPrice, currentPrice)*(1-pair.GetBuyDelta()), tickSizeExp)
			} else if lastExecutedSide == futures.SideTypeBuy {
				upPrice = round(math.Max(breakEvenPrice, currentPrice)*(1+pair.GetSellDelta()), tickSizeExp)
				delta := currentPrice - lastFilledPrice
				downPrice = round(math.Min(lastFilledPrice, currentPrice)*(1-pair.GetBuyDelta())-delta, tickSizeExp)
			}
			// Визначаємо ціну для нових ордерів коли позиція нульова
		} else {
			upPrice = round(currentPrice*(1+pair.GetSellDelta()), tickSizeExp)
			downPrice = round(currentPrice*(1-pair.GetBuyDelta()), tickSizeExp)
		}
	} else {
		upPrice = round(currentPrice*(1+pair.GetSellDelta()), tickSizeExp)
		downPrice = round(currentPrice*(1-pair.GetBuyDelta()), tickSizeExp)
	}
	return
}

func getQuantityPair(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	free float64,
	risk *futures.PositionRisk,
	minNotional,
	quantity,
	upPrice float64,
	stepSizeExp int,
	closeAll bool) (upQuantity, downQuantity float64) {
	var (
		freeNew float64
	)
	if config.GetConfigurations().GetDynamicQuantity() {
		if free*pair.GetLimitOnPosition() > pair.GetCurrentPositionBalance() {
			freeNew = pair.GetCurrentPositionBalance() * float64(pair.GetLeverage())
		} else {
			freeNew = free * pair.GetLimitOnPosition() * float64(pair.GetLeverage())
		}
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			upQuantity = round(freeNew*pair.GetLimitOnTransaction()/upPrice, stepSizeExp)
			if upQuantity*upPrice < minNotional {
				upQuantity = round(minNotional/upPrice, stepSizeExp)
			}
			if closeAll {
				downQuantity = utils.ConvStrToFloat64(risk.PositionAmt)
			} else {
				downQuantity = upQuantity
			}
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			downQuantity = round(freeNew*pair.GetLimitOnTransaction()/upPrice, stepSizeExp)
			if downQuantity*upPrice < minNotional {
				downQuantity = round(minNotional/upPrice, stepSizeExp)
			}
			if closeAll {
				upQuantity = utils.ConvStrToFloat64(risk.PositionAmt)
			} else {
				upQuantity = downQuantity
			}
		} else {
			upQuantity = quantity
			downQuantity = quantity
		}
	} else {
		upQuantity = quantity
		downQuantity = quantity
	}
	logrus.Debugf("Futures %s: PositionAmt %v, Free %v * LimitOnPosition %v = %v, CurrentPositionBalance %v",
		pair.GetPair(),
		utils.ConvStrToFloat64(risk.PositionAmt),
		free,
		pair.GetLimitOnPosition(),
		free*pair.GetLimitOnPosition(),
		pair.GetCurrentPositionBalance())
	logrus.Debugf("Futures %s: UpQuantity %v, DownQuantity %v", pair.GetPair(), upQuantity, downQuantity)
	return
}

func getCallBack_v3(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	quantity float64,
	tickSizeExp int,
	stepSizeExp int,
	minNotional float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			updateConfig(config, pair)
			// // Знаходимо у гріді на якому був виконаний ордер
			// if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
			// 	maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
			if event.OrderTradeUpdate.Type == futures.OrderTypeLimit {
				logrus.Debugf("Futures %s: Limited Order filled %v on price %v with quantity %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
			} else if event.OrderTradeUpdate.Type == futures.OrderTypeTakeProfitMarket {
				logrus.Debugf("Futures %s: Take Profit Order filled %v on price %v with quantity %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
			} else if event.OrderTradeUpdate.Type == futures.OrderTypeTrailingStopMarket {
				logrus.Debugf("Futures %s: Trailing Stop Order filled %v on price %v with quantity %v side %v status %s",
					pair.GetPair(),
					event.OrderTradeUpdate.ID,
					event.OrderTradeUpdate.OriginalPrice,
					event.OrderTradeUpdate.LastFilledQty,
					event.OrderTradeUpdate.Side,
					event.OrderTradeUpdate.Status)
			}
			free, _ := pairProcessor.GetFreeBalance()
			pair.SetCurrentBalance(free)
			config.Save()
			risk, err := pairProcessor.GetPositionRisk()
			if err != nil {
				printError()
				pairProcessor.CancelAllOrders()
				close(quit)
				return
			}
			// Визначаємо поточну ціну
			currentPrice := getCurrentPrice(client, pair, tickSizeExp)
			logrus.Debugf("Futures %s: Risks PositionAmt %v EntryPrice %v, BreakEvenPrice %v, Current Price %v, UnRealizedProfit %v",
				pair.GetPair(), risk.PositionAmt, risk.EntryPrice, risk.BreakEvenPrice, currentPrice, risk.UnRealizedProfit)
			logrus.Debugf("Futures %s: Event OrderTradeUpdate: OriginalPrice %v, OriginalQty %v, LastFilledPrice %v, LastFilledQty %v",
				pair.GetPair(), event.OrderTradeUpdate.OriginalPrice, event.OrderTradeUpdate.OriginalQty, event.OrderTradeUpdate.LastFilledPrice, event.OrderTradeUpdate.LastFilledQty)
			// Балансування маржі як треба
			free, _ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
			pairProcessor.CancelAllOrders()
			logrus.Debugf("Futures %s: Other orders was cancelled", pair.GetPair())
			err = createNextPair_v3(
				config,
				pair,
				risk,
				utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
				event.OrderTradeUpdate.Side,
				minNotional,
				quantity,
				currentPrice,
				tickSizeExp,
				stepSizeExp,
				free,
				pairProcessor)
			if err != nil {
				logrus.Errorf("Futures %s: %v", pair.GetPair(), err)
				printError()
				close(quit)
				return
			}
			// }
		}
	}
}

func createNextPair_v3(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	lastFilledPrice float64,
	lastExecutedSide futures.SideType,
	minNotional float64,
	quantity float64,
	currentPrice float64,
	tickSizeExp int,
	sizeSizeExp int,
	free float64,
	pairProcessor *PairProcessor) (err error) {
	var (
		upPrice          float64
		downPrice        float64
		upQuantity       float64
		downQuantity     float64
		createdOrderUp   bool = false
		createdOrderDown bool = false
		sellOrder        *futures.CreateOrderResponse
		buyOrder         *futures.CreateOrderResponse
	)
	// Визначаємо ціну для нових ордерів
	upPrice, downPrice = GetPricePair(
		config.GetConfigurations().GetDynamicDelta(),
		pair,
		lastFilledPrice,
		lastExecutedSide,
		risk, currentPrice, tickSizeExp)
	// Визначаємо кількість для нових ордерів
	upQuantity, downQuantity = getQuantityPair(config, pair, free, risk, minNotional, quantity, upPrice, sizeSizeExp, false)
	positionVal := utils.ConvStrToFloat64(risk.PositionAmt) * currentPrice / float64(pair.GetLeverage())
	// Створюємо ордер на продаж
	if pair.GetUpBound() != 0 && upPrice <= pair.GetUpBound() {
		if positionVal >= -pair.GetCurrentPositionBalance() {
			logrus.Debugf("Futures %s: Sell Quantity Up %v * upPrice %v = %v, minNotional %v",
				pair.GetPair(), upQuantity, upPrice, upQuantity*upPrice, minNotional)
			sellOrder, err = createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeLimit, upQuantity, upPrice, 0, false)
			if err != nil {
				printError()
				return
			}
			logrus.Debugf("Futures %s: Create Sell order on price %v quantity %v status %v",
				pair.GetPair(), upPrice, upQuantity, sellOrder.Status)
			if sellOrder.Status == futures.OrderStatusTypeFilled {
				pairProcessor.CancelAllOrders()
				risk, _ = pairProcessor.GetPositionRisk()
				return createNextPair_v3(
					config,
					pair,
					risk,
					upPrice,
					futures.SideTypeSell,
					minNotional,
					quantity,
					currentPrice,
					tickSizeExp,
					sizeSizeExp,
					free,
					pairProcessor)
			}
			createdOrderUp = true
		} else {
			logrus.Debugf("Futures %s: IsolatedMargin %v more than current position balance %v",
				pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
		}
	} else {
		logrus.Debugf("Futures %s: upPrice %v more than upBound %v",
			pair.GetPair(), upPrice, pair.GetUpBound())
	}
	// Створюємо ордер на купівлю
	if pair.GetLowBound() != 0 && downPrice >= pair.GetLowBound() {
		if positionVal <= pair.GetCurrentPositionBalance() {
			logrus.Debugf("Futures %s: Buy Quantity Down %v * downPrice %v = %v, minNotional %v",
				pair.GetPair(), downQuantity, downPrice, downQuantity*upPrice, minNotional)
			buyOrder, err = createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeLimit, downQuantity, downPrice, 0, false)
			if err != nil {
				printError()
				return
			}
			logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v status %v",
				pair.GetPair(), downPrice, downQuantity, buyOrder.Status)
			if buyOrder.Status == futures.OrderStatusTypeFilled {
				pairProcessor.CancelAllOrders()
				risk, _ = pairProcessor.GetPositionRisk()
				return createNextPair_v3(
					config,
					pair,
					risk,
					upPrice,
					futures.SideTypeSell,
					minNotional,
					quantity,
					currentPrice,
					tickSizeExp,
					sizeSizeExp,
					free,
					pairProcessor)
			}
			createdOrderDown = true
		} else {
			logrus.Debugf("Futures %s: IsolatedMargin %v more than current position balance %v",
				pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
		}
	} else {
		logrus.Debugf("Futures %s: downPrice %v less than lowBound %v",
			pair.GetPair(), downPrice, pair.GetLowBound())
	}
	if !config.GetConfigurations().GetObservePositionLoss() {
		if !createdOrderUp && !createdOrderDown {
			logrus.Debugf("Futures %s: Orders was not created", pair.GetPair())
			printError()
			return fmt.Errorf("orders were not created")
		}
	}
	return
}

func timeProcess(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quantity float64,
	risk *futures.PositionRisk,
	tickSizeExp int,
	pairProcessor *PairProcessor) (err error) {
	var (
		free         float64
		currentPrice float64
	)
	if config.GetConfigurations().GetBalancingOfMargin() ||
		config.GetConfigurations().GetObservePriceLiquidation() ||
		config.GetConfigurations().GetObservePositionLoss() {
		risk, err = pairProcessor.GetPositionRisk()
		if err != nil {
			return nil // Повертаємо nil, щоб продовжити роботу
		}
		free, _ = pairProcessor.GetFreeBalance()
		// Визначаємо поточну ціну
		if val, err := GetCurrentPrice(client, pair.GetPair()); err == nil { // Отримання ціни по ринку для пари
			currentPrice = round(val, tickSizeExp)
		} else {
			printError()
			return err
		}
	}
	// Балансування маржі як треба
	free, _ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
	// Обробка наближення ліквідаціі
	err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, currentPrice, quantity)
	if err != nil {
		printError()
		return err
	}
	// Обробка втрат по поточній позиції
	err = reOpenPosition(config, pair, risk, pairProcessor, quantity, currentPrice, tickSizeExp)
	if err != nil {
		printError()
		return err
	}
	return
}

// Працюємо лімітними ордерами (але можливо зменьшувати позицію будемо і TakeProfit ордером),
// відкриваємо ордера на продаж та купівлю з однаковою кількістью
// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
// після 5 наприклад ордера ставимо на більшу відстань
func RunFuturesGridTradingV3(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		initPriceUp   float64
		initPriceDown float64
		quantity      float64
		minNotional   float64
		tickSizeExp   int
		stepSizeExp   int
		pairProcessor *PairProcessor
		risk          *futures.PositionRisk
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV3)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairProcessor, err = initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		return err
	}
	_, _, initPriceUp, initPriceDown, quantity, minNotional, tickSizeExp, _, err = initVars(
		config.GetConfigurations().GetDynamicDelta(),
		client,
		pair,
		pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pair.GetCurrentPositionBalance(), pair.GetLimitOnTransaction())
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	_, err = streamStart(client, getCallBack_v3(
		config,
		client,
		pair,
		pairProcessor,
		quantity,
		tickSizeExp,
		stepSizeExp,
		minNotional,
		quit,
		maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	if config.GetConfigurations().GetObservePositionLoss() {
		go func() {
			for {
				<-time.After(time.Duration(config.GetConfigurations().GetObserverTimeOutMillisecond()) * time.Millisecond)
				risk, _ = pairProcessor.GetPositionRisk()
				if err != nil {
					continue // Спробуємо ще раз через ObserverTimeOutMillisecond
				}
				err = timeProcess(
					config,
					client,
					pair,
					quantity,
					risk,
					tickSizeExp,
					pairProcessor)
				if err != nil {
					pairProcessor.CancelAllOrders()
					printError()
					logrus.Errorf("Futures %s: Bot was stopped with error %v", pair.GetPair(), err)
					close(quit)
					return
				}
			}
		}()
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(pair, quantity, initPriceUp, initPriceDown, pairProcessor)
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		printError()
		return err
	}
	pairProcessor.CancelAllOrders()
	return nil
}

func streamStart(
	client *futures.Client,
	wsHandler func(*futures.WsUserDataEvent)) (resetEvent chan error, err error) {
	// Отримуємо ключ для прослуховування подій користувача
	listenKey, err := client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return
	}
	// Ініціалізуємо канал для відправки подій про необхідність оновлення стріму подій користувача
	resetEvent = make(chan error, 1)
	// Ініціалізуємо обробник помилок
	wsErrorHandler := func(err error) {
		resetEvent <- err
	}
	// Запускаємо стрім подій користувача
	_, _, err = futures.WsUserDataServe(listenKey, wsHandler, wsErrorHandler)
	return
}

func getCallBack_v4(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	pairProcessor *PairProcessor,
	quantity float64,
	tickSizeExp int,
	stepSizeExp int,
	minNotional float64,
	quit chan struct{},
	maintainedOrders *btree.BTree) func(*futures.WsUserDataEvent) {
	return func(event *futures.WsUserDataEvent) {
		if event.Event == futures.UserDataEventTypeOrderTradeUpdate &&
			event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
			updateConfig(config, pair)
			// Знаходимо у гріді на якому був виконаний ордер
			if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
				maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
				if event.OrderTradeUpdate.Type == futures.OrderTypeLimit {
					logrus.Debugf("Futures %s: Limited Order filled %v on price %v with quantity %v side %v status %s",
						pair.GetPair(),
						event.OrderTradeUpdate.ID,
						event.OrderTradeUpdate.OriginalPrice,
						event.OrderTradeUpdate.LastFilledQty,
						event.OrderTradeUpdate.Side,
						event.OrderTradeUpdate.Status)
				} else if event.OrderTradeUpdate.Type == futures.OrderTypeTakeProfitMarket {
					logrus.Debugf("Futures %s: Take Profit Order filled %v on price %v with quantity %v side %v status %s",
						pair.GetPair(),
						event.OrderTradeUpdate.ID,
						event.OrderTradeUpdate.OriginalPrice,
						event.OrderTradeUpdate.LastFilledQty,
						event.OrderTradeUpdate.Side,
						event.OrderTradeUpdate.Status)
				}
				free, _ := pairProcessor.GetFreeBalance()
				pair.SetCurrentBalance(free)
				config.Save()
				risk, err := pairProcessor.GetPositionRisk()
				if err != nil {
					printError()
					pairProcessor.CancelAllOrders()
					close(quit)
					return
				}
				// Визначаємо поточну ціну
				currentPrice := getCurrentPrice(client, pair, tickSizeExp)
				logrus.Debugf("Futures %s: Risks PositionAmt %v EntryPrice %v, BreakEvenPrice %v, Current Price %v, UnRealizedProfit %v",
					pair.GetPair(), risk.PositionAmt, risk.EntryPrice, risk.BreakEvenPrice, currentPrice, risk.UnRealizedProfit)
				logrus.Debugf("Futures %s: Event OrderTradeUpdate: OriginalPrice %v, OriginalQty %v, LastFilledPrice %v, LastFilledQty %v",
					pair.GetPair(), event.OrderTradeUpdate.OriginalPrice, event.OrderTradeUpdate.OriginalQty, event.OrderTradeUpdate.LastFilledPrice, event.OrderTradeUpdate.LastFilledQty)
				// Балансування маржі як треба
				free, _ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
				pairProcessor.CancelAllOrders()
				logrus.Debugf("Futures %s: Other orders was cancelled", pair.GetPair())
				err = createNextPair_v4(
					config,
					pair,
					risk,
					utils.ConvStrToFloat64(event.OrderTradeUpdate.LastFilledPrice),
					event.OrderTradeUpdate.Side,
					minNotional,
					quantity,
					currentPrice,
					tickSizeExp,
					stepSizeExp,
					free,
					pairProcessor)
				if err != nil {
					logrus.Errorf("Futures %s: %v", pair.GetPair(), err)
					printError()
					close(quit)
					return
				}
			}
		}
	}
}

func createNextPair_v4(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	lastFilledPrice float64,
	lastExecutedSide futures.SideType,
	minNotional float64,
	quantity float64,
	currentPrice float64,
	tickSizeExp int,
	sizeSizeExp int,
	free float64,
	pairProcessor *PairProcessor) (err error) {
	var (
		upPrice          float64
		downPrice        float64
		upQuantity       float64
		downQuantity     float64
		callbackRate     float64 = pair.GetCallbackRate()
		createdOrderUp   bool    = false
		createdOrderDown bool    = false
		sellOrder        *futures.CreateOrderResponse
		buyOrder         *futures.CreateOrderResponse
	)
	// Визначаємо ціну для нових ордерів
	upPrice, downPrice = GetPricePair(
		config.GetConfigurations().GetDynamicDelta(),
		pair,
		lastFilledPrice,
		lastExecutedSide,
		risk,
		currentPrice,
		tickSizeExp)
	getClosePosition := func(risk *futures.PositionRisk) (up, down bool) {
		// Визначаємо кількість для нових ордерів коли позиція від'ємна
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			up = false
			down = true
			// Визначаємо кількість для нових ордерів коли позиція позитивна
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			up = true
			down = false
			// Визначаємо кількість для нових ордерів коли позиція нульова
		} else {
			up = false
			down = false
		}
		return
	}
	// Визначаємо кількість для нових ордерів
	upQuantity, downQuantity = getQuantityPair(config, pair, free, risk, minNotional, quantity, upPrice, sizeSizeExp, true)
	upClosePosition, downClosePosition := getClosePosition(risk)
	if pair.GetUpBound() != 0 && upPrice <= pair.GetUpBound() && upQuantity > 0 {
		if upClosePosition {
			sellOrder, err = createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeTrailingStopMarket, upQuantity, upPrice, callbackRate, upClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Try to create Sell order on price %v quantity %v minNotional/upPrice %v",
					pair.GetPair(), downPrice, downQuantity, minNotional/upPrice)
				printError()
				return
			}
		} else {
			sellOrder, err = createOrder(pairProcessor, futures.SideTypeSell, futures.OrderTypeLimit, upQuantity, upPrice, 0, upClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Try to create Sell order on price %v quantity %v minNotional/upPrice %v",
					pair.GetPair(), downPrice, downQuantity, minNotional/upPrice)
				printError()
				return
			}
		}
		logrus.Debugf("Futures %s: Create Sell order on price %v quantity %v status %v",
			pair.GetPair(), upPrice, upQuantity, sellOrder.Status)
		if sellOrder.Status == futures.OrderStatusTypeFilled {
			pairProcessor.CancelAllOrders()
			risk, _ = pairProcessor.GetPositionRisk()
			return createNextPair_v4(
				config,
				pair,
				risk,
				upPrice,
				futures.SideTypeSell,
				minNotional,
				quantity,
				currentPrice,
				tickSizeExp,
				sizeSizeExp,
				free,
				pairProcessor)
		}
		createdOrderUp = true
	} else {
		if upQuantity <= 0 {
			logrus.Debugf("Futures %s: upQuantity %v less than 0", pair.GetPair(), upQuantity)
		} else {
			logrus.Debugf("Futures %s: upPrice %v more than upBound %v",
				pair.GetPair(), upPrice, pair.GetUpBound())
		}
	}
	// Створюємо ордер на купівлю
	if pair.GetLowBound() != 0 && downPrice >= pair.GetLowBound() && downQuantity > 0 {
		if downClosePosition {
			buyOrder, err = createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeTrailingStopMarket, downQuantity, downPrice, callbackRate, downClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Try to create Buy order on price %v quantity %v minNotional/upPrice %v",
					pair.GetPair(), downPrice, downQuantity, minNotional/upPrice)
				printError()
				return
			}
		} else {
			buyOrder, err = createOrder(pairProcessor, futures.SideTypeBuy, futures.OrderTypeLimit, downQuantity, downPrice, 0, downClosePosition)
			if err != nil {
				logrus.Errorf("Futures %s: Try to create Buy order on price %v quantity %v minNotional/upPrice %v",
					pair.GetPair(), downPrice, downQuantity, minNotional/upPrice)
				printError()
				return
			}
		}
		logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v status %v",
			pair.GetPair(), downPrice, downQuantity, buyOrder.Status)
		if buyOrder.Status == futures.OrderStatusTypeFilled {
			pairProcessor.CancelAllOrders()
			risk, _ = pairProcessor.GetPositionRisk()
			return createNextPair_v4(
				config,
				pair,
				risk,
				upPrice,
				futures.SideTypeSell,
				minNotional,
				quantity,
				currentPrice,
				tickSizeExp,
				sizeSizeExp,
				free,
				pairProcessor)
		}
		createdOrderDown = true
		logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v", pair.GetPair(), downPrice, downQuantity)
	} else {
		if downQuantity <= 0 {
			logrus.Debugf("Futures %s: downQuantity %v less than 0", pair.GetPair(), downQuantity)
		} else {
			logrus.Debugf("Futures %s: downPrice %v less than lowBound %v",
				pair.GetPair(), downPrice, pair.GetLowBound())
		}
	}
	if !createdOrderUp && !createdOrderDown {
		logrus.Debugf("Futures %s: Orders was not created", pair.GetPair())
		printError()
		return fmt.Errorf("orders were not created")
	}
	return
}

// Працюємо лімітними та TakeProfit/TrailingStop ордерами,
// відкриваємо лімітний ордер на збільшення, а закриваємо всю позицію TakeProfit/TrailingStop або лімітним ордером
// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
// після 5 наприклад ордера ставимо на більшу відстань
// або статично відкриємо ордери на продаж та купівлю з однаковою кількістью та с однаковим шагом
func RunFuturesGridTradingV4(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		initPriceUp   float64
		initPriceDown float64
		quantity      float64
		minNotional   float64
		tickSizeExp   int
		stepSizeExp   int
		pairProcessor *PairProcessor
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV5)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairProcessor, err = initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		return err
	}

	_, _, initPriceUp, initPriceDown, quantity, minNotional, tickSizeExp, _, err = initVars(
		config.GetConfigurations().GetDynamicDelta(),
		client,
		pair,
		pairProcessor)
	if err != nil {
		return err
	}
	if minNotional > pair.GetCurrentPositionBalance()*pair.GetLimitOnTransaction() {
		printError()
		return fmt.Errorf("minNotional %v more than current position balance %v * limitOnTransaction %v",
			minNotional, pair.GetCurrentPositionBalance(), pair.GetLimitOnTransaction())
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	_, err = streamStart(client, getCallBack_v4(config, client, pair, pairProcessor, quantity, tickSizeExp, stepSizeExp, minNotional, quit, maintainedOrders))
	if err != nil {
		printError()
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = openPosition(pair, quantity, initPriceUp, initPriceDown, pairProcessor)
	if err != nil {
		return err
	}
	<-quit
	logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
	err = loadConfig(pair, config, pairProcessor)
	if err != nil {
		printError()
		return err
	}
	pairProcessor.CancelAllOrders()
	return nil
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	quit chan struct{},
	debug bool,
	wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		var err error
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			err = fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			err = RunFuturesHolding(config, client, degree, limit, pair, quit, time.Second, debug, wg)

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			err = RunScalpingHolding(config, client, pair, quit, wg)

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			err = RunFuturesTrading(config, client, degree, limit, pair, quit, time.Second, debug, wg)

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			err = RunFuturesGridTrading(config, client, pair, quit, wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV2 {
			err = RunFuturesGridTradingV2(config, client, pair, quit, wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV3 {
			err = RunFuturesGridTradingV3(config, client, pair, quit, wg)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV4 {
			err = RunFuturesGridTradingV4(config, client, pair, quit, wg)

			// } else if pair.GetStrategy() == pairs_types.GridStrategyTypeV5 {
			// 	err = RunFuturesGridTradingV5(config, client, pair, quit, wg)

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			err = fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
		if err != nil {
			logrus.Error(err)
			close(quit)
		}
	}()
}
