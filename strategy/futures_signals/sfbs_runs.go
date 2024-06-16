package futures_signals

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	// "github.com/fr0ster/go-trading-utils/binance/futures/account"

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
			downOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, order.GetDownPrice())
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
			upOrder, err := createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, order.GetUpPrice())
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

func loadConfig(pair *pairs_types.Pairs, config *config_types.ConfigFile, pairStreams *PairStreams) (err error) {
	baseValue, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
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
	quit chan struct{}) (pairStreams *PairStreams, pairProcessor *PairProcessor, err error) {
	// Створюємо стрім подій
	pairStreams, err = NewPairStreams(client, pair, quit, false)
	if err != nil {
		printError()
		return
	}
	// Створюємо обробник пари
	pairProcessor, err = NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), quit, false)
	if err != nil {
		printError()
		return
	}

	balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
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
	pairStreams *PairStreams) (res *futures.Symbol, err error) {
	val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()})
	if val == nil {
		printError()
		return nil, fmt.Errorf("futures %s: Symbol not found", pair.GetPair())
	}
	return val.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
}

func initVars(
	client *futures.Client,
	pair *pairs_types.Pairs,
	pairStreams *PairStreams) (
	symbol *futures.Symbol,
	price,
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
	symbol, err = getSymbol(pair, pairStreams)
	if err != nil {
		printError()
		return
	}
	tickSizeExp = getTickSizeExp(symbol)
	stepSizeExp = getStepSizeExp(symbol)
	// Отримання середньої ціни
	price = round(pair.GetMiddlePrice(), tickSizeExp)
	if price <= 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = round(price, tickSizeExp)
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

func initFirstPairOfOrders(
	pair *pairs_types.Pairs,
	price float64,
	quantity float64,
	tickSizeExp int,
	pairProcessor *PairProcessor) (sellOrder, buyOrder *futures.CreateOrderResponse, err error) {
	err = pairProcessor.CancelAllOrders()
	if err != nil {
		printError()
		return
	}

	// Створюємо ордери на продаж
	sellOrder, err = createOrderInGrid(pairProcessor, futures.SideTypeSell, quantity, round(price*(1+pair.GetSellDelta()), tickSizeExp))
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set Sell order on price %v with quantity %v", pair.GetPair(), round(price*(1+pair.GetSellDelta()), tickSizeExp), quantity)
	// Створюємо ордери на купівлю
	buyOrder, err = createOrderInGrid(pairProcessor, futures.SideTypeBuy, quantity, round(price*(1-pair.GetBuyDelta()), tickSizeExp))
	if err != nil {
		printError()
		return
	}
	logrus.Debugf("Futures %s: Set Buy order on price %v with quantity %v", pair.GetPair(), round(price*(1-pair.GetBuyDelta()), tickSizeExp), quantity)
	return
}

func getCurrentPrice(
	client *futures.Client,
	pair *pairs_types.Pairs,
	tickSizeExp int) (currentPrice float64) {
	val, _ := GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
	currentPrice = round(val, tickSizeExp)
	return
}

func marginBalancing(
	config *config_types.ConfigFile,
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	pairProcessor *PairProcessor,
	free float64,
	tickStepSize int) (err error) {
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

func positionLossObservation(
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
				_, _, err = initFirstPairOfOrders(pair, price, quantity, tickSizeExp, pairProcessor)
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

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		quantity     float64
		locked       float64
		free         float64
		currentPrice float64
		risk         *futures.PositionRisk
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyType)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairStreams, pairProcessor, err := initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairStreams)
	if err != nil {
		return err
	}
	symbol, initPrice, quantity, _, tickSizeExp, _, err := initVars(client, pair, pairStreams)
	if err != nil {
		return err
	}
	sellOrder, buyOrder, err := initFirstPairOfOrders(pair, currentPrice, quantity, tickSizeExp, pairProcessor)
	if err != nil {
		return err
	}
	// Ініціалізація гріду
	grid, err := initGrid(pair, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	for {
		select {
		case <-quit:
			err = loadConfig(pair, config, pairStreams)
			if err != nil {
				printError()
				return err
			}
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
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
					if !ok {
						if !(event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled) {
							return fmt.Errorf("uncorrected order ID: %v", event.OrderTradeUpdate.ID)
						} else {
							continue // Вважаємо ордер обробили раніше???
						}
					}
					orderId := order.GetOrderId()
					locked, _ = pairStreams.GetAccount().GetLockedAsset(pair.GetBaseSymbol())
					free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
					pair.SetCurrentBalance(free)
					config.Save()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						grid.Unlock()
						printError()
						return
					}
					// Балансування маржі як треба
					_ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
					// Обробка наближення ліквідаціі
					err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, initPrice, quantity)
					if err != nil {
						grid.Unlock()
						return err
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
						grid.Unlock()
						pairProcessor.CancelAllOrders()
						printError()
						return err
					}
					grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
					grid.Unlock()
				}
			}
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
		quantity     float64
		locked       float64
		free         float64
		currentPrice float64
		risk         *futures.PositionRisk
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV2)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairStreams, pairProcessor, err := initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairStreams)
	if err != nil {
		return err
	}
	symbol, initPrice, quantity, _, tickSizeExp, _, err := initVars(client, pair, pairStreams)
	if err != nil {
		return err
	}
	sellOrder, buyOrder, err := initFirstPairOfOrders(pair, currentPrice, quantity, tickSizeExp, pairProcessor)
	if err != nil {
		return err
	}
	// Ініціалізація гріду
	grid, err := initGrid(pair, initPrice, quantity, tickSizeExp, sellOrder, buyOrder)
	maintainedOrders := btree.New(2)
	if err != nil {
		printError()
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	for {
		select {
		case <-quit:
			err = loadConfig(pair, config, pairStreams)
			if err != nil {
				printError()
				return err
			}
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			grid.Lock()
			updateConfig(config, pair)
			// Знаходимо у гріді на якому був виконаний ордер
			currentPrice = utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)
			order, ok := grid.Get(&grid_types.Record{Price: currentPrice}).(*grid_types.Record)
			if !ok {
				printError()
				return fmt.Errorf("we didn't work with order on price level %v before: %v", currentPrice, event.OrderTradeUpdate.ID)
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
				locked, _ = pairStreams.GetAccount().GetLockedAsset(pair.GetBaseSymbol())
				free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
				pair.SetCurrentBalance(free)
				config.Save()
				risk, err = pairProcessor.GetPositionRisk()
				if err != nil {
					grid.Unlock()
					printError()
					return
				}
				// Балансування маржі як треба
				_ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
				// Обробка наближення ліквідаціі
				err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, initPrice, quantity)
				if err != nil {
					grid.Unlock()
					return err
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
					grid.Unlock()
					pairProcessor.CancelAllOrders()
					printError()
					return err
				}
				grid.Debug("Futures Grid processOrder", strconv.FormatInt(orderId, 10), pair.GetPair())
			}
			grid.Unlock()
		}
	}
}

func getQuantity(
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	currentPrice float64,
	quantity float64,
	minNotional float64,
	stepSizeExp int,
	positionLimit float64) (correctedQuantityUp, correctedQuantityDown, positionVal float64) {
	positionVal = utils.ConvStrToFloat64(risk.PositionAmt) * currentPrice / float64(pair.GetLeverage())
	minQuantity := round(minNotional/currentPrice, stepSizeExp)
	// Коефіцієнт кількості в одному ордері відносно поточного балансу та позиції
	quantityCoefficient := (positionLimit - math.Abs(positionVal)) / positionLimit
	if positionVal < 0 {
		// Якщо позиція від'ємна, то зменшуємо кількість на новому ордері на продаж на коефіцієнт коррекції
		if quantity > minQuantity {
			correctedQuantityUp = round((quantity-minQuantity)*quantityCoefficient, stepSizeExp) + minQuantity
		} else {
			correctedQuantityUp = minQuantity
		}
		// Але кількість на купівлю не змінюємо
		if quantity > minQuantity {
			correctedQuantityDown = quantity
		} else {
			correctedQuantityDown = minQuantity
		}
	} else if positionVal > 0 {
		// Якщо позиція позитивна,кількість на продаж не змінюємо
		if quantity > minQuantity {
			correctedQuantityUp = quantity
		} else {
			correctedQuantityUp = minQuantity
		}
		// Але зменшуємо кількість на новому ордері на купівлю на коефіцієнт коррекції
		if quantity > minQuantity {
			correctedQuantityDown = round((quantity-minQuantity)*quantityCoefficient, stepSizeExp) + minQuantity
		} else {
			correctedQuantityDown = minQuantity
		}
	} else {
		// Якщо позиція нульова, то кількість на продаж та купівлю однакова
		correctedQuantityUp = quantity
		correctedQuantityDown = quantity
	}
	logrus.Debugf("Futures %s: (Position limit %v - math.Abs(positionVal) %v) %v / Position limit %v = QuantityCoefficient %v",
		pair.GetPair(),
		positionLimit,
		math.Abs(positionVal),
		(positionLimit - math.Abs(positionVal)),
		positionLimit,
		quantityCoefficient)
	return
}

func getPrice(
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	currentPrice float64,
	tickSizeExp int) (upPrice, downPrice float64) {
	breakEvenPrice := utils.ConvStrToFloat64(risk.BreakEvenPrice)
	entryPrice := utils.ConvStrToFloat64(risk.EntryPrice)
	// Визначаємо ціну для нових ордерів коли позиція від'ємна
	if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
		upPrice = round(entryPrice*(1+pair.GetSellDelta()), tickSizeExp)
		downPrice = round(breakEvenPrice*(1-pair.GetBuyDelta()), tickSizeExp)
		// Визначаємо ціну для нових ордерів коли позиція позитивна
	} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
		upPrice = round(breakEvenPrice*(1+pair.GetSellDelta()), tickSizeExp)
		downPrice = round(entryPrice*(1-pair.GetBuyDelta()), tickSizeExp)
		// Визначаємо ціну для нових ордерів коли позиція нульова
	} else {
		upPrice = round(currentPrice*(1+pair.GetSellDelta()), tickSizeExp)
		downPrice = round(currentPrice*(1-pair.GetBuyDelta()), tickSizeExp)
	}
	return
}

func createNextPair_v0(
	pair *pairs_types.Pairs,
	currentPrice float64,
	quantity float64,
	minNotional float64,
	risk *futures.PositionRisk,
	tickSizeExp int,
	stepSizeExp int,
	positionLimit float64,
	pairProcessor *PairProcessor) (err error) {
	var (
		correctedQuantityUp   float64
		correctedQuantityDown float64
	)
	// Визначаємо кількість для нових ордерів
	correctedQuantityUp, correctedQuantityDown, positionVal := getQuantity(
		pair,
		risk,
		currentPrice,
		quantity,
		minNotional,
		stepSizeExp,
		positionLimit)
	// Визначаємо ціну для нових ордерів
	upPrice, downPrice := getPrice(pair, risk, currentPrice, tickSizeExp)
	// Створюємо ордер на продаж
	if pair.GetUpBound() != 0 && upPrice <= pair.GetUpBound() {
		if positionVal >= -pair.GetCurrentPositionBalance() {
			logrus.Debugf("Futures %s: Corrected Quantity Up %v * upPrice %v = %v, minNotional %v",
				pair.GetPair(), correctedQuantityUp, upPrice, correctedQuantityUp*upPrice, minNotional)
			_, err = createOrderInGrid(pairProcessor, futures.SideTypeSell, correctedQuantityUp, upPrice)
			if err != nil {
				printError()
				return
			}
			logrus.Debugf("Futures %s: Create Sell order on price %v quantity %v", pair.GetPair(), upPrice, correctedQuantityUp)
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
			logrus.Debugf("Futures %s: Corrected Quantity Down %v * downPrice %v = %v, minNotional %v",
				pair.GetPair(), correctedQuantityDown, downPrice, correctedQuantityDown*upPrice, minNotional)
			_, err = createOrderInGrid(pairProcessor, futures.SideTypeBuy, correctedQuantityDown, downPrice)
			if err != nil {
				printError()
				return
			}
			logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v", pair.GetPair(), downPrice, correctedQuantityDown)
		} else {
			logrus.Debugf("Futures %s: IsolatedMargin %v more than current position balance %v",
				pair.GetPair(), risk.IsolatedMargin, pair.GetCurrentPositionBalance())
		}
	} else {
		logrus.Debugf("Futures %s: downPrice %v less than lowBound %v",
			pair.GetPair(), downPrice, pair.GetLowBound())
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
	pairStreams *PairStreams,
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
			printError()
			return
		}
		free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
		// Визначаємо поточну ціну
		if val, err := GetPrice(client, pair.GetPair()); err == nil { // Отримання ціни по ринку для пари
			currentPrice = round(val, tickSizeExp)
		} else {
			printError()
			return err
		}
	}
	// Балансування маржі як треба
	_ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
	// Обробка наближення ліквідаціі
	err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, currentPrice, quantity)
	if err != nil {
		return err
	}
	// Обробка втрат по поточній позиції
	err = positionLossObservation(config, pair, risk, pairProcessor, quantity, currentPrice, tickSizeExp)
	if err != nil {
		return err
	}
	return
}

func RunFuturesGridTradingV3(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		initPrice     float64
		quantity      float64
		free          float64
		currentPrice  float64
		minNotional   float64
		tickSizeExp   int
		stepSizeExp   int
		pairStreams   *PairStreams
		pairProcessor *PairProcessor
		risk          *futures.PositionRisk
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV3)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairStreams, pairProcessor, err = initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairStreams)
	if err != nil {
		return err
	}
	_, initPrice, quantity, minNotional, tickSizeExp, stepSizeExp, err = initVars(client, pair, pairStreams)
	if err != nil {
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = initFirstPairOfOrders(pair, initPrice, quantity, tickSizeExp, pairProcessor)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	for {
		select {
		case <-quit:
			err = loadConfig(pair, config, pairStreams)
			if err != nil {
				printError()
				return err
			}
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
				updateConfig(config, pair)
				// Знаходимо у гріді на якому був виконаний ордер
				if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
					maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
					logrus.Debugf("Futures %s: Order filled %v on price %v with quantity %v side %v status %s",
						pair.GetPair(),
						event.OrderTradeUpdate.ID,
						event.OrderTradeUpdate.OriginalPrice,
						event.OrderTradeUpdate.LastFilledQty,
						event.OrderTradeUpdate.Side,
						event.OrderTradeUpdate.Status)
					free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
					pair.SetCurrentBalance(free)
					config.Save()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						printError()
						pairProcessor.CancelAllOrders()
						return
					}
					logrus.Debugf("Futures %s: Risks EntryPrice %v, BreakEvenPrice %v, Current Price %v, UnRealizedProfit %v",
						pair.GetPair(), risk.EntryPrice, risk.BreakEvenPrice, currentPrice, risk.UnRealizedProfit)
					// Визначаємо поточну ціну
					currentPrice = getCurrentPrice(client, pair, tickSizeExp)
					// Балансування маржі як треба
					_ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
					// Обробка наближення ліквідаціі
					err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, initPrice, quantity)
					if err != nil {
						printError()
						pairProcessor.CancelAllOrders()
						return err
					}
					pairProcessor.CancelAllOrders()
					logrus.Debugf("Futures %s: Other orders was cancelled", pair.GetPair())
					err = createNextPair_v0(
						pair,
						currentPrice,
						quantity,
						minNotional,
						risk,
						tickSizeExp,
						stepSizeExp,
						pair.GetCurrentPositionBalance(),
						pairProcessor)
					if err != nil {
						pairProcessor.CancelAllOrders()
						return err
					}
				}
			}
		case <-time.After(time.Duration(config.GetConfigurations().GetObserverTimeOutMillisecond()) * time.Millisecond):
			risk, err = pairProcessor.GetPositionRisk()
			if err != nil {
				printError()
				pairProcessor.CancelAllOrders()
				return
			}
			err = timeProcess(
				config,
				client,
				pair,
				quantity,
				risk,
				tickSizeExp,
				pairStreams,
				pairProcessor)
			if err != nil {
				pairProcessor.CancelAllOrders()
				return err
			}
		}
	}
}

func createNextPair_v1(
	pair *pairs_types.Pairs,
	risk *futures.PositionRisk,
	currentPrice float64,
	quantity float64,
	tickSizeExp int,
	pairProcessor *PairProcessor) (err error) {
	var (
		upPrice      float64
		downPrice    float64
		upQuantity   float64
		downQuantity float64
	)
	getQuantity := func(risk *futures.PositionRisk) (upQuantity, downQuantity float64) {
		// Визначаємо кількість для нових ордерів коли позиція від'ємна
		if utils.ConvStrToFloat64(risk.PositionAmt) < 0 {
			upQuantity = quantity
			downQuantity = utils.ConvStrToFloat64(risk.PositionAmt) * -1
			// Визначаємо кількість для нових ордерів коли позиція позитивна
		} else if utils.ConvStrToFloat64(risk.PositionAmt) > 0 {
			upQuantity = utils.ConvStrToFloat64(risk.PositionAmt)
			downQuantity = quantity
			// Визначаємо кількість для нових ордерів коли позиція нульова
		} else {
			upQuantity = quantity
			downQuantity = quantity
		}
		return
	}
	upPrice, downPrice = getPrice(pair, risk, currentPrice, tickSizeExp)
	upQuantity, downQuantity = getQuantity(risk)
	if pair.GetUpBound() != 0 && upPrice <= pair.GetUpBound() {
		_, err = createOrderInGrid(pairProcessor, futures.SideTypeSell, upQuantity, upPrice)
		if err != nil {
			printError()
			return
		}
		logrus.Debugf("Futures %s: Create Sell order on price %v quantity %v", pair.GetPair(), upPrice, quantity)
	} else {
		logrus.Debugf("Futures %s: upPrice %v more than upBound %v",
			pair.GetPair(), upPrice, pair.GetUpBound())
	}
	// Створюємо ордер на купівлю
	if pair.GetLowBound() != 0 && downPrice >= pair.GetLowBound() {
		_, err = createOrderInGrid(pairProcessor, futures.SideTypeBuy, downQuantity, downPrice)
		if err != nil {
			printError()
			return
		}
		logrus.Debugf("Futures %s: Create Buy order on price %v quantity %v", pair.GetPair(), downPrice, quantity)
	} else {
		logrus.Debugf("Futures %s: downPrice %v less than lowBound %v",
			pair.GetPair(), downPrice, pair.GetLowBound())
	}
	return
}

func RunFuturesGridTradingV4(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair *pairs_types.Pairs,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	var (
		initPrice     float64
		quantity      float64
		free          float64
		currentPrice  float64
		tickSizeExp   int
		pairStreams   *PairStreams
		pairProcessor *PairProcessor
		risk          *futures.PositionRisk
	)
	err = checkRun(pair, pairs_types.USDTFutureType, pairs_types.GridStrategyTypeV4)
	if err != nil {
		return err
	}
	// Створюємо стрім подій
	pairStreams, pairProcessor, err = initRun(config, client, pair, quit)
	if err != nil {
		return err
	}
	err = loadConfig(pair, config, pairStreams)
	if err != nil {
		return err
	}
	_, initPrice, quantity, _, tickSizeExp, _, err = initVars(client, pair, pairStreams)
	if err != nil {
		return err
	}
	// Створюємо початкові ордери на продаж та купівлю
	_, _, err = initFirstPairOfOrders(pair, initPrice, quantity, tickSizeExp, pairProcessor)
	if err != nil {
		return err
	}
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	maintainedOrders := btree.New(2)
	for {
		select {
		case <-quit:
			err = loadConfig(pair, config, pairStreams)
			if err != nil {
				printError()
				return err
			}
			pairProcessor.CancelAllOrders()
			logrus.Infof("Futures %s: Bot was stopped", pair.GetPair())
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			if event.OrderTradeUpdate.Status == futures.OrderStatusTypeFilled {
				updateConfig(config, pair)
				// Знаходимо у гріді на якому був виконаний ордер
				if !maintainedOrders.Has(grid_types.OrderIdType(event.OrderTradeUpdate.ID)) {
					maintainedOrders.ReplaceOrInsert(grid_types.OrderIdType(event.OrderTradeUpdate.ID))
					logrus.Debugf("Futures %s: Order filled %v on price %v with quantity %v side %v status %s",
						pair.GetPair(),
						event.OrderTradeUpdate.ID,
						event.OrderTradeUpdate.OriginalPrice,
						event.OrderTradeUpdate.LastFilledQty,
						event.OrderTradeUpdate.Side,
						event.OrderTradeUpdate.Status)
					free, _ = pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
					pair.SetCurrentBalance(free)
					config.Save()
					risk, err = pairProcessor.GetPositionRisk()
					if err != nil {
						printError()
						pairProcessor.CancelAllOrders()
						return
					}
					// Визначаємо поточну ціну
					currentPrice = getCurrentPrice(client, pair, tickSizeExp)
					logrus.Debugf("Futures %s: Risks EntryPrice %v, BreakEvenPrice %v, Current Price %v, UnRealizedProfit %v",
						pair.GetPair(), risk.EntryPrice, risk.BreakEvenPrice, currentPrice, risk.UnRealizedProfit)
					// Балансування маржі як треба
					_ = marginBalancing(config, pair, risk, pairProcessor, free, tickSizeExp)
					// Обробка наближення ліквідаціі
					err = liquidationObservation(config, pair, risk, pairProcessor, currentPrice, free, initPrice, quantity)
					if err != nil {
						printError()
						pairProcessor.CancelAllOrders()
						return err
					}
					pairProcessor.CancelAllOrders()
					logrus.Debugf("Futures %s: Other orders was cancelled", pair.GetPair())
					err = createNextPair_v1(
						pair,
						risk,
						currentPrice,
						quantity,
						tickSizeExp,
						pairProcessor)
					if err != nil {
						pairProcessor.CancelAllOrders()
						return err
					}
				}
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
