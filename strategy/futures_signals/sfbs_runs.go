package futures_signals

import (
	"fmt"
	"os"
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
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
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
	if pair.GetStrategy() != pairs_types.TradingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}
	stopEvent <- os.Interrupt
	return fmt.Errorf("it hadn't been implemented yet")
}

func RunFuturesGridTrading(
	config *config_types.ConfigFile,
	client *futures.Client,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
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
	getPosition := func() *futures.PositionRisk {
		risks, err := pairStreams.GetAccount().GetPositionRisk(pair.GetPair())
		if err != nil {
			logrus.Errorf("Futures %s: %v\n", pair.GetPair(), err)
			stopEvent <- os.Interrupt
			return nil
		}
		if len(risks) != 0 {
			return risks[0]
		} else {
			logrus.Errorf("Futures %s: Position not found\n", pair.GetPair())
			stopEvent <- os.Interrupt
			return nil
		}
	}
	// Ініціалізація гріду
	logrus.Debugf("Futures %s: Grid initialized\n", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	// Отримання середньої ціни
	price := pair.GetMiddlePrice()
	if utils.ConvStrToFloat64(getPosition().EntryPrice) != 0 {
		price = utils.ConvStrToFloat64(getPosition().EntryPrice)
	}
	if price == 0 {
		stopEvent <- os.Interrupt
		return fmt.Errorf("futures %s: We don`t have position and didn`t set Middle Price (we use it as EntryPrice) in Config", pair.GetPair())
	}
	quantity := utils.ConvStrToFloat64(getPosition().PositionAmt) * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction()
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
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), false)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Створюємо ордери на продаж
	sellOrder, err := pairProcessor.CreateOrder(
		futures.OrderTypeLimit,        // orderType
		futures.SideTypeSell,          // sideType
		futures.TimeInForceTypeGTC,    // timeInForce
		quantity,                      // quantity
		false,                         // closePosition
		price*(1+pair.GetSellDelta()), // price
		0,                             // stopPrice
		0)                             // trailingDelta
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, price*(1+pair.GetSellDelta()), price, 0, types.SideTypeSell))
	// Створюємо ордер на купівлю
	buyOrder, err := pairProcessor.CreateOrder(
		futures.OrderTypeLimit,       // orderType
		futures.SideTypeBuy,          // sideType
		futures.TimeInForceTypeGTC,   // timeInForce
		quantity,                     // quantity
		false,                        // closePosition
		price*(1-pair.GetBuyDelta()), // price
		0,                            // stopPrice
		0)                            // trailingDelta
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, price*(1-pair.GetBuyDelta()), 0, price, types.SideTypeBuy))
	// Стартуємо обробку ордерів
	logrus.Debugf("Futures %s: Start Order Status Event\n", pair.GetPair())
	event := pairProcessor.GetOrderStatusEvent()
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return nil
		case event := <-event:
			var (
				upOrder   *grid_types.Record
				downOrder *grid_types.Record
			)
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{OrderId: event.OrderTradeUpdate.ID}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v\n", event.OrderTradeUpdate.ID)
				continue
			}
			logrus.Debugf("Futures %s: Read Order by ID %v from grid\n", pair.GetPair(), event.OrderTradeUpdate.ID)
			upOrder, ok = grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
			if !ok {
				if pair.GetUpBound() != 0 && price*(1+pair.GetSellDelta()) > pair.GetUpBound() {
					continue
				}
				upOrder = grid_types.NewRecord(0, price*(1+pair.GetSellDelta()), price, 0, types.SideTypeSell)
				grid.Set(upOrder)
			}
			logrus.Debugf("Futures %s: Read Up Order by price %v from grid\n", pair.GetPair(), order.GetPrice())
			downOrder, ok = grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
			if !ok {
				if pair.GetLowBound() != 0 && price*(1-pair.GetBuyDelta()) > pair.GetLowBound() {
					continue
				}
				downOrder = grid_types.NewRecord(0, price*(1-pair.GetBuyDelta()), 0, price, types.SideTypeBuy)
				grid.Set(downOrder)
			}
			logrus.Debugf("Futures %s: Read Low Order by price %v from grid\n", pair.GetPair(), order.GetPrice())
			if upOrder.GetOrderId() == 0 || downOrder.GetOrderId() == 0 {
				logrus.Warnf("Futures %s: Order on price below and above hadn't been filled yet\n", pair.GetPair())
				continue
			}
			// Виконаний ордер помічаємо як виконаний
			logrus.Debugf("Futures %s: Executed Order %v marked as Filled\n", pair.GetPair(), order.GetOrderId())
			order.SetOrderId(0)
			order.SetOrderSide(types.SideTypeNone)
			// Створюємо нові ордери
			// Якщо на шабель вище ордер не розміщено , то створюємо ордер на продаж
			if upOrder.GetOrderId() == 0 {
				logrus.Debugf("Futures %s: Sell order on price %v\n", pair.GetPair(), upOrder.GetUpPrice())
				sellOrder, err := pairProcessor.CreateOrder(
					futures.OrderTypeLimit,     // orderType
					futures.SideTypeSell,       // sideType
					futures.TimeInForceTypeGTC, // timeInForce
					quantity,                   // quantity
					false,                      // closePosition
					upOrder.GetPrice(),         // price
					0,                          // stopPrice
					0)                          // trailingDelta
				if err != nil {
					stopEvent <- os.Interrupt
					return err
				}
				upOrder.SetOrderId(sellOrder.OrderID)
				upOrder.SetOrderSide(types.SideTypeSell)
			}
			// Якщо на шабель нижче ордер не розміщено , то створюємо ордер на купівлю
			if downOrder.GetOrderId() == 0 {
				logrus.Debugf("Futures %s: Buy order on price %v\n", pair.GetPair(), downOrder.GetDownPrice())
				buyOrder, err := pairProcessor.CreateOrder(
					futures.OrderTypeLimit,     // orderType
					futures.SideTypeBuy,        // sideType
					futures.TimeInForceTypeGTC, // timeInForce
					quantity,                   // quantity
					false,                      // closePosition
					downOrder.GetPrice(),       // price
					0,                          // stopPrice
					0)                          // trailingDelta
				if err != nil {
					stopEvent <- os.Interrupt
					return err
				}
				downOrder.SetOrderId(buyOrder.OrderID)
				downOrder.SetOrderSide(types.SideTypeBuy)
			}
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
