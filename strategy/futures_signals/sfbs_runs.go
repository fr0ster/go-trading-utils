package futures_signals

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	types "github.com/fr0ster/go-trading-utils/types"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	grid_types "github.com/fr0ster/go-trading-utils/types/grid"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

const (
	deltaUp   = 0.0005
	deltaDown = 0.0005
	degree    = 3
	limit     = 1000
	interval  = "1m"
)

func RunFuturesHolding(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
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
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal) (err error) {
	pair.SetStrategy(pairs_types.GridStrategyType)
	return RunFuturesGridTrading(config, client, pair, stopEvent)
}

func RunFuturesTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
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
	stopEvent <- os.Interrupt
	return fmt.Errorf("it hadn't been implemented yet")
}

func getPositionRisk(pairStreams *PairStreams, pair pairs_interfaces.Pairs) (risks *futures.PositionRisk, err error) {
	risks, err = pairStreams.GetAccount().GetPositionRisk(pair.GetPair())
	return
}

// Створення ордера для розміщення в грід
func initOrderInGrid(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair pairs_interfaces.Pairs,
	side futures.SideType,
	quantity,
	price float64) (order *futures.CreateOrderResponse, err error) {
	for {
		order, err := pairProcessor.CreateOrder(
			futures.OrderTypeLimit,     // orderType
			side,                       // sideType
			futures.TimeInForceTypeGTC, // timeInForce
			quantity,                   // quantity
			false,                      // closePosition
			price,                      // price
			0,                          // stopPrice
			0)                          // callbackRate
		if err != nil {
			return nil, err
		}
		if order.Status != futures.OrderStatusTypeNew {
			pair.SetBuyDelta(pair.GetBuyDelta() * 2)
			pair.SetSellDelta(pair.GetSellDelta() * 2)
			config.Save()
		} else {
			return order, nil
		}
	}
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair pairs_interfaces.Pairs,
	side futures.SideType,
	grid *grid_types.Grid,
	quantity float64,
	existNextPrice,
	nonExistNextPrice float64,
	stopEvent chan os.Signal) (err error) {
	var (
		nextOrder *grid_types.Record
		ok        bool
	)
	if existNextPrice != 0 { // Якщо запис вище існує ...
		nextOrder, ok = grid.Get(&grid_types.Record{Price: existNextPrice}).(*grid_types.Record)
		if ok {
			if nextOrder.GetOrderId() == 0 { // ... і він не має ID ордера
				// Створюємо ордер на продаж
				sellOrder, err := initOrderInGrid(config, pairProcessor, pair, side, quantity, existNextPrice)
				if err != nil {
					stopEvent <- os.Interrupt
					return err
				}
				// Записуємо номер ордера в грід
				nextOrder.SetOrderId(sellOrder.OrderID)
				grid.Set(nextOrder)
				nextOrder.SetOrderSide(types.OrderSide(side))
				if side == futures.SideTypeBuy {
					logrus.Debugf("Futures %s: Set Buy order %v on price %v", pair.GetPair(), nextOrder.GetOrderId(), existNextPrice)
				} else {
					logrus.Debugf("Futures %s: Set Sell order %v on price %v", pair.GetPair(), nextOrder.GetOrderId(), existNextPrice)
				}
			} else {
				stopEvent <- os.Interrupt
				return fmt.Errorf("futures %s: Order on price above hadn't been filled yet", pair.GetPair())
			}
		}
	} else { // Якщо запис вище не існує
		// Створюємо ордер на продаж
		nextOrder, err := initOrderInGrid(config, pairProcessor, pair, side, quantity, nonExistNextPrice)
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		// Записуємо ордер в грід
		grid.Set(grid_types.NewRecord(nextOrder.OrderID, nonExistNextPrice, 0, nonExistNextPrice, types.OrderSide(side)))
		if side == futures.SideTypeBuy {
			logrus.Debugf("Futures %s: Add Buy order %v on price %v", pair.GetPair(), nextOrder.OrderID, nonExistNextPrice)
		} else {
			logrus.Debugf("Futures %s: Add Sell order %v on price %v", pair.GetPair(), nextOrder.OrderID, nonExistNextPrice)
		}
	}
	return
}

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal) (err error) {
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
		return
	}
	if pair.GetInitialBalance() == 0 {
		balance, err := pairStreams.GetAccount().GetFreeAsset(pair.GetBaseSymbol())
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		pair.SetInitialBalance(balance)
		config.Save()
	}
	if pair.GetInitialPositionBalance() == 0 {
		pair.SetInitialPositionBalance(pair.GetInitialBalance() * pair.GetLimitOnPosition())
		config.Save()
	}
	if pair.GetCurrentBalance() == 0 {
		pair.SetCurrentBalance(pair.GetInitialBalance())
		config.Save()
	}
	if pair.GetCurrentPositionBalance() == 0 {
		pair.SetCurrentPositionBalance(pair.GetInitialPositionBalance())
		config.Save()
	}
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	// Отримання середньої ціни
	price := pair.GetMiddlePrice()
	risk, err := getPositionRisk(pairStreams, pair)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	if entryPrice := utils.ConvStrToFloat64(risk.EntryPrice); entryPrice != 0 {
		price = entryPrice
	}
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
	}
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() / price
	if symbol := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()}); symbol != nil {
		val, err := symbol.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		minNotional := utils.ConvStrToFloat64(val.MinNotionalFilter().Notional)
		if quantity*price < minNotional {
			quantity = minNotional / price
		}
	} else {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: Symbol not found", pair.GetPair())
	}
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, price*(1+pair.GetSellDelta()), price*(1-pair.GetBuyDelta()), types.SideTypeNone))
	logrus.Debugf("Futures %s: Set Entry Price order on price %v", pair.GetPair(), price)
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), false)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Створюємо ордери на продаж
	sellOrder, err := initOrderInGrid(config, pairProcessor, pair, futures.SideTypeSell, quantity, price*(1+pair.GetSellDelta()))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, price*(1+pair.GetSellDelta()), 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), price*(1+pair.GetSellDelta()))
	// Створюємо ордер на купівлю
	buyOrder, err := initOrderInGrid(config, pairProcessor, pair, futures.SideTypeBuy, quantity, price*(1-pair.GetSellDelta()))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, price*(1-pair.GetSellDelta()), price, 0, types.SideTypeBuy))
	logrus.Debugf("Futures %s: Set Buy order on price %v", pair.GetPair(), price*(1-pair.GetBuyDelta()))
	// Стартуємо обробку ордерів
	grid.Debug("Futures Grid", pair.GetPair())
	logrus.Debugf("Futures %s: Start Order Status Event", pair.GetPair())
	mu := &sync.Mutex{}
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			mu.Lock()
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{OrderId: event.OrderTradeUpdate.ID}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v", event.OrderTradeUpdate.ID)
				mu.Unlock()
				continue
			}
			order.SetOrderId(0)                    // Помічаємо ордер як виконаний
			order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
			// Якшо куплено цільової валюти більше ніж потрібно, то не робимо новий ордер
			if val, err := getPositionRisk(pairStreams, pair); err == nil && math.Abs(utils.ConvStrToFloat64(val.PositionAmt)*utils.ConvStrToFloat64(val.EntryPrice)) > pair.GetCurrentBalance()*pair.GetLimitOnPosition() {
				logrus.Debugf("Futures %s: Target value %v above limit %v", pair.GetPair(), pair.GetBuyQuantity()*pair.GetBuyValue()-pair.GetSellQuantity()*pair.GetSellValue(), pair.GetCurrentBalance()*pair.GetLimitOnPosition())
				mu.Unlock()
				continue
			} else if err != nil {
				mu.Unlock()
				stopEvent <- os.Interrupt
				return err
			}
			logrus.Debugf("Futures %s: Read Order by ID %v from grid", pair.GetPair(), event.OrderTradeUpdate.ID)
			if pair.GetUpBound() != 0 && order.GetUpPrice() > pair.GetUpBound() {
				logrus.Debugf("Futures %s: Price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
				mu.Unlock()
				continue
			}
			err = processOrder(config, pairProcessor, pair, futures.SideTypeSell, grid, quantity, order.GetUpPrice(), order.GetPrice()*(1+pair.GetSellDelta()), stopEvent)
			if err != nil {
				mu.Unlock()
				stopEvent <- os.Interrupt
				return err
			}
			if pair.GetLowBound() != 0 && order.GetDownPrice() < pair.GetLowBound() {
				logrus.Debugf("Futures %s: Price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
				mu.Unlock()
				continue
			}
			err = processOrder(config, pairProcessor, pair, futures.SideTypeBuy, grid, quantity, order.GetDownPrice(), order.GetPrice()*(1-pair.GetSellDelta()), stopEvent)
			if err != nil {
				mu.Unlock()
				stopEvent <- os.Interrupt
				return err
			}
			mu.Unlock()
		case <-time.After(60 * time.Second):
			grid.Debug("Futures Grid", pair.GetPair())
		}
	}
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	debug bool) (err error) {
	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		stopEvent <- os.Interrupt
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
		stopEvent <- os.Interrupt
		return fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
	}
}
