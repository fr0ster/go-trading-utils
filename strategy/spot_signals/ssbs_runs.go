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

// Створення ордера для розміщення в грід
func initOrderInGrid(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair pairs_interfaces.Pairs,
	side binance.SideType,
	quantity,
	price float64) (order *binance.CreateOrderResponse, err error) {
	for {
		order, err := pairProcessor.CreateOrder(
			binance.OrderTypeLimit,     // orderType
			side,                       // sideType
			binance.TimeInForceTypeGTC, // timeInForce
			quantity,                   // quantity
			0,                          // quantityQty
			price,                      // price
			0,                          // stopPrice
			0)                          // trailingDelta
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

// Обробка ордерів після виконання ордера з гріду
func processOrder(
	config *config_types.ConfigFile,
	pairProcessor *PairProcessor,
	pair pairs_interfaces.Pairs,
	side binance.SideType,
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
				nextOrder.SetOrderSide(types.OrderSide(side))
				grid.Set(nextOrder)
				if side == binance.SideTypeBuy {
					logrus.Debugf("Spot %s: Set Buy order %v on price %v", pair.GetPair(), nextOrder.GetOrderId(), existNextPrice)
				} else {
					logrus.Debugf("Spot %s: Set Sell order %v on price %v", pair.GetPair(), nextOrder.GetOrderId(), existNextPrice)
				}
			} else {
				stopEvent <- os.Interrupt
				return fmt.Errorf("spot %s: Order on price above hadn't been filled yet", pair.GetPair())
			}
		}
	} else { // Якщо запис вище не існує
		// Створюємо ордер на продаж
		sellOrder, err := initOrderInGrid(config, pairProcessor, pair, side, quantity, nonExistNextPrice)
		if err != nil {
			stopEvent <- os.Interrupt
			return err
		}
		// Записуємо ордер в грід
		grid.Set(grid_types.NewRecord(sellOrder.OrderID, nonExistNextPrice, 0, nonExistNextPrice, types.OrderSide(side)))
		if side == binance.SideTypeBuy {
			logrus.Debugf("Spot %s: Add Buy order %v on price %v", pair.GetPair(), sellOrder.OrderID, nonExistNextPrice)
		} else {
			logrus.Debugf("Spot %s: Add Sell order %v on price %v", pair.GetPair(), nextOrder.GetOrderId(), nonExistNextPrice)
		}
	}
	return
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
	quantity := pair.GetCurrentBalance() * pair.GetLimitOnPosition() * pair.GetLimitOnTransaction() / price
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
	// Створюємо ордер на продаж
	sellOrder, err := initOrderInGrid(config, pairProcessor, pair, binance.SideTypeSell, quantity, price*(1+pair.GetSellDelta()))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(sellOrder.OrderID, price*(1+pair.GetSellDelta()), 0, price, types.SideTypeSell))
	logrus.Debugf("Spot %s: Set Sell order on price %v", pair.GetPair(), price*(1+pair.GetSellDelta()))
	// Створюємо ордер на купівлю
	buyOrder, err := initOrderInGrid(config, pairProcessor, pair, binance.SideTypeBuy, quantity, price*(1-pair.GetSellDelta()))
	if err != nil {
		stopEvent <- os.Interrupt
		return err
	}
	// Записуємо ордер в грід
	grid.Set(grid_types.NewRecord(buyOrder.OrderID, price*(1+pair.GetSellDelta()), price, 0, types.SideTypeBuy))
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
			logrus.Debugf("Spot %s: Order %v status %s", pair.GetPair(), event.OrderUpdate.Id, event.OrderUpdate.Status)
			// Знаходимо у гріді відповідний запис, та записи на шабель вище та нижче
			order, ok := grid.Get(&grid_types.Record{OrderId: event.OrderUpdate.Id}).(*grid_types.Record)
			if !ok {
				logrus.Errorf("Uncorrected order ID: %v", event.OrderUpdate.Id)
				continue
			}
			order.SetOrderId(0)                    // Помічаємо ордер як виконаний
			order.SetOrderSide(types.SideTypeNone) // Помічаємо ордер як виконаний
			if pair.GetUpBound() != 0 && order.GetUpPrice() > pair.GetUpBound() {
				logrus.Debugf("Spot %s: Price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
				continue
			}
			err = processOrder(config, pairProcessor, pair, binance.SideTypeSell, grid, quantity, order.GetUpPrice(), order.GetPrice()*(1+pair.GetSellDelta()), stopEvent)
			if err != nil {
				stopEvent <- os.Interrupt
				return err
			}
			if pair.GetLowBound() != 0 && order.GetDownPrice() < pair.GetLowBound() {
				logrus.Debugf("Spot %s: Price %v above upper bound %v", pair.GetPair(), price*(1+pair.GetSellDelta()), pair.GetUpBound())
				continue
			}
			err = processOrder(config, pairProcessor, pair, binance.SideTypeBuy, grid, quantity, order.GetDownPrice(), order.GetPrice()*(1-pair.GetSellDelta()), stopEvent)
			if err != nil {
				stopEvent <- os.Interrupt
				return err
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
