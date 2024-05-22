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

// Округлення ціни до TickSize знаків після коми
func roundPrice(val float64, symbol *futures.Symbol) float64 {
	exp := int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(symbol.PriceFilter().TickSize)))))
	return utils.RoundToDecimalPlace(val, exp)
}

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair pairs_interfaces.Pairs,
	symbol *futures.Symbol,
	side futures.SideType,
	grid *grid_types.Grid,
	order *grid_types.Record,
	quantity float64) (err error) {
	var (
		// Стара середня точка, розміщуємо ордер на продаж чи купівлю
		oldPrice *grid_types.Record
		oldOrder *futures.CreateOrderResponse
		// Нова середня точка, перевіряємо, чи існує відповідний запис у grid,
		// якшо не існує та новий запис у межах діапазону,
		// то створюємо нову запис та розміщуємо ордер
		newOrder *futures.CreateOrderResponse
		ok       bool
	)
	if (side == futures.SideTypeSell && order.GetDownPrice() != 0) || (side == futures.SideTypeBuy && order.GetUpPrice() != 0) { // Якщо наступний запис існує ...
		if side == futures.SideTypeSell {
			oldPrice, ok = grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
		} else {
			oldPrice, ok = grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
		}
		if ok {
			if oldPrice.GetOrderId() == 0 { // ... і він не має ID ордера
				// Створюємо ордер на продаж чи купівлю
				if side == futures.SideTypeSell {
					oldOrder, err = initOrderInGrid(config, pairProcessor, pair, side, quantity, order.GetUpPrice())
				} else {
					oldOrder, err = initOrderInGrid(config, pairProcessor, pair, side, quantity, order.GetDownPrice())
				}
				if err != nil {
					return err
				}
				// Записуємо номер ордера в грід
				oldPrice.SetOrderId(oldOrder.OrderID)
				grid.Set(oldPrice)
				oldPrice.SetOrderSide(types.OrderSide(side))
				if side == futures.SideTypeBuy {
					logrus.Debugf("Futures %s: Set Buy order %v on price %v", pair.GetPair(), oldPrice.GetOrderId(), order.GetUpPrice())
				} else {
					logrus.Debugf("Futures %s: Set Sell order %v on price %v", pair.GetPair(), oldPrice.GetOrderId(), order.GetDownPrice())
				}
			} else {
				return fmt.Errorf("futures %s: Order on price above hadn't been filled yet", pair.GetPair())
			}
		}
	}
	if (side == futures.SideTypeSell && order.GetUpPrice() == 0) || (side == futures.SideTypeBuy && order.GetDownPrice() == 0) { // Якщо запис вище не існує
		if side == futures.SideTypeSell {
			// Створюємо ордер на продаж
			price := roundPrice(order.GetPrice()*(1+pair.GetSellDelta()), symbol)
			if price > pair.GetUpBound() {
				return fmt.Errorf("futures %s: Price %v above up bound %v", pair.GetPair(), price, pair.GetUpBound())
			}
			newOrder, err = initOrderInGrid(config, pairProcessor, pair, side, quantity, price)
			// Записуємо ордер в грід
			grid.Set(grid_types.NewRecord(newOrder.OrderID, price, 0, order.GetPrice(), types.OrderSide(side)))
			logrus.Debugf("Futures %s: Add Sell order %v on price %v", pair.GetPair(), newOrder.OrderID, price)
		} else {
			// Створюємо ордер на продаж
			price := roundPrice(order.GetPrice()*(1-pair.GetBuyDelta()), symbol)
			if price < pair.GetLowBound() {
				return fmt.Errorf("futures %s: Price %v below down bound %v", pair.GetPair(), price, pair.GetLowBound())
			}
			newOrder, err = initOrderInGrid(config, pairProcessor, pair, side, quantity, price)
			// Записуємо ордер в грід
			grid.Set(grid_types.NewRecord(newOrder.OrderID, price, order.GetPrice(), 0, types.OrderSide(side)))
			logrus.Debugf("Futures %s: Add Buy order %v on price %v", pair.GetPair(), newOrder.OrderID, price)
		}
		if err != nil {
			return err
		}
	}
	order.SetOrderId(0)                    // Помічаємо ордер як виконаний
	order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
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
	symbol, err := func() (res *futures.Symbol, err error) {
		val := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair.GetPair()})
		if val == nil {
			return nil, fmt.Errorf("spot %s: Symbol not found", pair.GetPair())
		}
		return val.(*symbol_info.FuturesSymbol).GetFuturesSymbol()
	}()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Отримання середньої ціни
	price := roundPrice(pair.GetMiddlePrice(), symbol)
	risk, err := getPositionRisk(pairStreams, pair)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	if entryPrice := utils.ConvStrToFloat64(risk.EntryPrice); entryPrice != 0 {
		price = entryPrice // Ціна позиції вже округлена до symbol.LotSizeFilter().StepSize
	}
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
		price = roundPrice(price, symbol)
	}
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() / price
	minNotional := utils.ConvStrToFloat64(symbol.MinNotionalFilter().Notional)
	if quantity*price < minNotional {
		quantity = minNotional / price
	}
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, roundPrice(price*(1+pair.GetSellDelta()), symbol), roundPrice(price*(1-pair.GetBuyDelta()), symbol), types.SideTypeNone))
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
	sellOrder, err := initOrderInGrid(config, pairProcessor, pair, futures.SideTypeSell, quantity, roundPrice(price*(1+pair.GetSellDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, roundPrice(price*(1+pair.GetSellDelta()), symbol), 0, price, types.SideTypeSell))
	logrus.Debugf("Futures %s: Set Sell order on price %v", pair.GetPair(), roundPrice(price*(1+pair.GetSellDelta()), symbol))
	// Створюємо ордер на купівлю
	buyOrder, err := initOrderInGrid(config, pairProcessor, pair, futures.SideTypeBuy, quantity, roundPrice(price*(1-pair.GetBuyDelta()), symbol))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, roundPrice(price*(1-pair.GetSellDelta()), symbol), price, 0, types.SideTypeBuy))
	logrus.Debugf("Futures %s: Set Buy order on price %v", pair.GetPair(), roundPrice(price*(1-pair.GetBuyDelta()), symbol))
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
			logrus.Debugf("Spot %s: Order %v status %s", pair.GetPair(), event.OrderTradeUpdate.ID, event.OrderTradeUpdate.Status)
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{Price: utils.ConvStrToFloat64(event.OrderTradeUpdate.OriginalPrice)}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v", event.OrderTradeUpdate.ID)
				mu.Unlock()
				continue
			}
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
			if event.OrderTradeUpdate.Side == futures.SideTypeBuy {
				err = processOrder(config, pairProcessor, pair, symbol, futures.SideTypeSell, grid, order, quantity)
				if err != nil {
					mu.Unlock()
					stopEvent <- os.Interrupt
					return err
				}
			} else {
				err = processOrder(config, pairProcessor, pair, symbol, futures.SideTypeBuy, grid, order, quantity)
				if err != nil {
					mu.Unlock()
					stopEvent <- os.Interrupt
					return err
				}
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
