package spot_signals

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
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

	pairProcessor, err := NewPairProcessor(config, client, pair, debug)
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

	pairProcessor, err := NewPairProcessor(config, client, pair, debug)
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

	pairProcessor, err := NewPairProcessor(config, client, pair, debug)
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
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	debug bool) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.GridStrategyType {
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

	pairProcessor, err := NewPairProcessor(config, client, pair, debug)
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
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		RunSpotHolding(
			config,
			client,
			degree,
			limit,
			pair,
			stopEvent,
			updateTime,
			debug)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		RunSpotScalping(
			config,
			client,
			degree,
			limit,
			pair,
			stopEvent,
			updateTime,
			debug)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {

		RunSpotTrading(
			config,
			client,
			degree,
			limit,
			pair,
			stopEvent,
			updateTime,
			debug)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
	}
	return nil
}
