package spot_signals

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
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
)

func RunSpotHolding(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage || pair.GetStage() == pairs_types.OutputOfPositionStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	buyEvent, _ := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)

	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-time.After(updateTime):
				triggerEvent <- true
			}
		}
	}()

	collectionOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	_, err = pairProcessor.ProcessBuyOrder(buyEvent)
	if err != nil {
		return err
	}

	<-collectionOutEvent
	pairProcessor.StopBuySignal()
	pair.SetStage(pairs_types.PositionClosedStage)
	config.Save()
	stopEvent <- os.Interrupt
	return nil
}

func RunSpotScalping(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessSellOrder(sellEvent) // Все одно другий раз не запустится, бо вже працює горутина
		if err != nil {
			return err
		}
		workingOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)
		_, err = pairProcessor.ProcessSellOrder(sellEvent)
		if err != nil {
			return err
		}

		<-workingOutEvent
		pairProcessor.StopBuySignal()
		pair.SetStage(pairs_types.OutputOfPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		if err != nil {
			return err
		}
		positionClosed := pairObserver.ClosePositionSignal(triggerEvent) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
	}
	return nil
}

func RunSpotTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.TradingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}
	if pair.GetStage() == pairs_types.PositionClosedStage {
		return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
	}

	account, err := spot_account.New(client, []string{pair.GetBaseSymbol(), pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairBookTickerObserver, err := NewPairBookTickersObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}

	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-buyEvent:
				triggerEvent <- true
			case <-sellEvent:
				triggerEvent <- true
			}
		}
	}()

	pairStream, err := NewPairStreams(client, pair, debug)
	if err != nil {
		return err
	}
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStream.GetExchangeInfo(), pairStream.GetAccount(), pairStream.GetUserDataEvent(), debug)
	if err != nil {
		return err
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeMarket) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeMarket)
	}

	if !pairProcessor.CheckOrderType(binance.OrderTypeTakeProfit) {
		return fmt.Errorf("pair %v has wrong order type %v", pair.GetPair(), binance.OrderTypeTakeProfitLimit)
	}

	if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
		_, err = pairProcessor.ProcessBuyOrder(buyEvent)
		if err != nil {
			return err
		}
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
		<-collectionOutEvent
		pair.SetStage(pairs_types.OutputOfPositionStage) // В trading стратегії не спекулюємо, накопили позицію і закриваемо продажем лімітним ордером
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		pairProcessor.StopBuySignal() // Зупиняємо купівлю, продаємо поки є шо продавати
		// TODO: Закриття позиції лімітним trailing ордером
		quantity, err := GetTargetBalance(account, pair)
		if err != nil {
			return err
		}
		order, err := pairProcessor.CreateOrder(
			binance.OrderTypeTakeProfitLimit,
			binance.SideTypeSell,
			binance.TimeInForceTypeGTC,
			// STOP_LOSS_LIMIT/TAKE_PROFIT_LIMIT timeInForce, quantity, price, stopPrice or trailingDelta
			quantity,
			0,   // quantityQty
			0,   // price
			0,   // stopPrice
			100) // trailingDelta
		if err != nil {
			return err
		}
		positionClosed := pairProcessor.OrderExecutionGuard(order) // Чекаємо на закриття позиції
		<-positionClosed
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
	}
	return nil
}

func RunSpotGridTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
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
		return err
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
	if pair.GetSellQuantity() == 0 && pair.GetBuyQuantity() == 0 {
		targetValue, err := pairStreams.GetAccount().GetFreeAsset(pair.GetTargetSymbol())
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		pair.SetBuyQuantity(targetValue)
		config.Save()
	}
	// Ініціалізація гріду
	logrus.Debugf("Spot %s: Grid initialized", pair.GetPair())
	grid := grid_types.New()
	// Перевірка на коректність дельт
	if pair.GetSellDelta() != pair.GetBuyDelta() {
		stopEvent <- os.Interrupt
		return fmt.Errorf("spot %s: SellDelta %v != BuyDelta %v", pair.GetPair(), pair.GetSellDelta(), pair.GetBuyDelta())
	}
	// Отримання середньої ціни
	price := pair.GetMiddlePrice()
	if price == 0 {
		price, _ = GetPrice(client, pair.GetPair()) // Отримання ціни по ринку для пари
	}
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction()
	if symbol := pairStreams.GetExchangeInfo().GetSymbol(&symbol_info.SpotSymbol{Symbol: pair.GetPair()}); symbol != nil {
		val, err := symbol.(*symbol_info.SpotSymbol).GetSpotSymbol()
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		minNotional := utils.ConvStrToFloat64(val.NotionalFilter().MinNotional)
		if quantity*price < minNotional {
			quantity = minNotional / price
		}
	} else {
		stopEvent <- os.Interrupt
		return fmt.Errorf("spot %s: Symbol not found", pair.GetPair())
	}
	// Записуємо середню ціну в грід
	grid.Set(grid_types.NewRecord(0, price, price*(1+pair.GetSellDelta()), price*(1-pair.GetBuyDelta()), types.SideTypeNone))
	logrus.Debugf("Spot %s: Set Entry Price order on price %v", pair.GetPair(), price)
	pairProcessor, err := NewPairProcessor(config, client, pair, pairStreams.GetExchangeInfo(), pairStreams.GetAccount(), pairStreams.GetUserDataEvent(), false)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	_, err = pairProcessor.CancelAllOrders()
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	initOrderInGrid := func(side binance.SideType, quantity float64) (order *binance.CreateOrderResponse, err error) {
		for {
			order, err := pairProcessor.CreateOrder(
				binance.OrderTypeLimit,        // orderType
				side,                          // sideType
				binance.TimeInForceTypeGTC,    // timeInForce
				quantity,                      // quantity
				0,                             // quantityQty
				price*(1+pair.GetSellDelta()), // price
				0,                             // stopPrice
				0)                             // trailingDelta
			if err != nil {
				return nil, err
			}
			if order.Status != binance.OrderStatusTypeNew {
				pair.SetBuyDelta(pair.GetBuyDelta() * 2)
				pair.SetSellDelta(pair.GetSellDelta() * 2)
				config.Save()
			} else {
				return order, nil
			}
		}
	}
	// Створюємо ордери на продаж
	sellOrder, err := initOrderInGrid(binance.SideTypeSell, quantity)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, price*(1+pair.GetSellDelta()), price, 0, types.SideTypeSell))
	logrus.Debugf("Spot %s: Set Sell order on price %v", pair.GetPair(), price*(1+pair.GetSellDelta()))
	// Створюємо ордер на купівлю
	buyOrder, err := initOrderInGrid(binance.SideTypeBuy, quantity)
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, price*(1+pair.GetSellDelta()), 0, price, types.SideTypeBuy))
	logrus.Debugf("Spot %s: Set Buy order on price %v", pair.GetPair(), price*(1-pair.GetBuyDelta()))
	// Стартуємо обробку ордерів
	grid.Debug("Spots Grid", pair.GetPair())
	logrus.Debugf("Spot %s: Start Order Processing", pair.GetPair())
	for {
		select {
		case <-stopEvent:
			stopEvent <- os.Interrupt
			return nil
		case event := <-pairProcessor.GetOrderStatusEvent():
			var (
				upOrder  *grid_types.Record
				lowOrder *grid_types.Record
			)
			logrus.Debugf("Spot %s: Order %v status %s", pair.GetPair(), event.OrderUpdate.Id, event.OrderUpdate.Status)
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{OrderId: event.OrderUpdate.Id}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v", event.OrderUpdate.Id)
				continue
			}
			logrus.Debugf("Spot %s: Read Order by ID %v from grid", pair.GetPair(), event.OrderUpdate.Id)
			upOrder, ok = grid.Get(&grid_types.Record{Price: order.GetUpPrice()}).(*grid_types.Record)
			if !ok {
				if pair.GetUpBound() != 0 && price*(1+pair.GetSellDelta()) > pair.GetUpBound() {
					logrus.Debugf("Spot %s: Price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
					continue
				}
				logrus.Debugf("Spot %s: Add order at price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
				upOrder = grid_types.NewRecord(0, price*(1+pair.GetSellDelta()), price, 0, types.SideTypeSell)
				grid.Set(upOrder)
			}
			logrus.Debugf("Spot %s: Read Up Order by price %v from grid", pair.GetPair(), order.GetPrice())
			lowOrder, ok = grid.Get(&grid_types.Record{Price: order.GetDownPrice()}).(*grid_types.Record)
			if !ok {
				if pair.GetLowBound() != 0 && price*(1-pair.GetBuyDelta()) > pair.GetLowBound() {
					logrus.Debugf("Spot %s: Price %v below lower bound %v", pair.GetPair(), price*(1-pair.GetBuyDelta()), pair.GetLowBound())
					continue
				}
				// Якшо куплено цільової валюти більше ніж потрібно, то не робимо новий ордер
				if (pair.GetBuyQuantity()*pair.GetBuyValue() - pair.GetSellQuantity()*pair.GetSellValue()) > pair.GetCurrentBalance()*pair.GetLimitOnPosition() {
					logrus.Debugf("Spot %s: Target value %v above limit %v", pair.GetPair(), pair.GetBuyQuantity()*pair.GetBuyValue()-pair.GetSellQuantity()*pair.GetSellValue(), pair.GetCurrentBalance()*pair.GetLimitOnPosition())
					continue
				}
				logrus.Debugf("Spot %s: Add order at price %v below lower bound %v\n", pair.GetPair(), price*(1-pair.GetBuyDelta()), pair.GetLowBound())
				lowOrder = grid_types.NewRecord(0, price*(1-pair.GetBuyDelta()), 0, price, types.SideTypeBuy)
				grid.Set(lowOrder)
			}
			logrus.Debugf("Spot %s: Read Low Order by price %v from grid", pair.GetPair(), order.GetPrice())
			if upOrder.GetOrderId() == 0 || lowOrder.GetOrderId() == 0 {
				logrus.Warnf("Spot %s: Order on price below and above hadn't been filled yet\n", pair.GetPair())
				continue
			}
			// Виконаний ордер помічаємо як виконаний
			logrus.Debugf("Spot %s: Executed Order %v marked as Filled", pair.GetPair(), order.GetOrderId())
			order.SetOrderId(0)
			order.SetOrderSide(types.SideTypeNone)
			// Створюємо нові ордери
			// Якщо на шабель вище ордер не розміщено , то створюємо ордер на продаж
			if upOrder.GetOrderId() == 0 {
				logrus.Debugf("Spot %s: Set Sell order on price %v", pair.GetPair(), upOrder.GetUpPrice())
				sellOrder, err := pairProcessor.CreateOrder(
					binance.OrderTypeLimit,     // orderType
					binance.SideTypeSell,       // sideType
					binance.TimeInForceTypeGTC, // timeInForce
					quantity,                   // quantity
					0,                          // quantityQty
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
			if lowOrder.GetOrderId() == 0 {
				logrus.Debugf("Spot %s: Set Buy order on price %v", pair.GetPair(), lowOrder.GetDownPrice())
				buyOrder, err := pairProcessor.CreateOrder(
					binance.OrderTypeLimit,     // orderType
					binance.SideTypeBuy,        // sideType
					binance.TimeInForceTypeGTC, // timeInForce
					quantity,                   // quantity
					0,                          // quantityQty
					lowOrder.GetPrice(),        // price
					0,                          // stopPrice
					0)                          // trailingDelta
				if err != nil {
					stopEvent <- os.Interrupt
					return err
				}
				lowOrder.SetOrderId(buyOrder.OrderID)
				lowOrder.SetOrderSide(types.SideTypeBuy)
			}
		case <-time.After(60 * time.Second):
			grid.Debug("Spots Grid", pair.GetPair())
		}
	}
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		stopEvent <- os.Interrupt
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		return RunSpotHolding(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		return RunSpotScalping(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		return RunSpotTrading(config, client, degree, limit, pair, stopEvent, updateTime, debug)

		// Відпрацьовуємо Grid стратегію
	} else if pair.GetStrategy() == pairs_types.GridStrategyType {
		return RunSpotGridTrading(config, client, pair, stopEvent)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		stopEvent <- os.Interrupt
		return fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
	}
}
