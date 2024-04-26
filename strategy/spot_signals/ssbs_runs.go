package spot_signals

import (
	"fmt"
	_ "net/http/pprof"
	"time"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"

	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
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
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.HoldingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver.StartBookTickersUpdateGuard()
	buyEvent, _ := pairObserver.StartBuyOrSellByBookTickerSignal()

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

	pairProcessor, err :=
		NewPairProcessor(
			config, client, pair, binance.OrderTypeMarket, buyEvent, nil)
	if err != nil {
		return err
	}

	_ = pairProcessor.ProcessBuyOrder()

	<-collectionOutEvent
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
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.ScalpingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver.StartBookTickersUpdateGuard()
	buyEvent, sellEvent := pairObserver.StartBuyOrSellByDepthSignal()

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

	pairProcessor, err :=
		NewPairProcessor(
			config, client, pair, binance.OrderTypeMarket, buyEvent, sellEvent)
	if err != nil {
		return err
	}

	_ = pairProcessor.ProcessBuyOrder()

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.WorkInPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.WorkInPositionStage {
		workingOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)
		_ = pairProcessor.ProcessSellOrder()

		<-workingOutEvent
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
	}
	return nil
}

func RunSpotTrading(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
	if pair.GetAccountType() != pairs_types.SpotAccountType {
		return fmt.Errorf("pair %v has wrong account type %v", pair.GetPair(), pair.GetAccountType())
	}
	if pair.GetStrategy() != pairs_types.TradingStrategyType {
		return fmt.Errorf("pair %v has wrong strategy %v", pair.GetPair(), pair.GetStrategy())
	}

	err = PairInit(client, config, account, pair)
	if err != nil {
		return err
	}

	RunConfigSaver(config, stopEvent, updateTime)

	pairObserver, err := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	if err != nil {
		return err
	}
	pairObserver.StartBookTickersUpdateGuard()
	buyEvent, sellEvent := pairObserver.StartBuyOrSellByBookTickerSignal()

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

	pairProcessor, err :=
		NewPairProcessor(
			config, client, pair, binance.OrderTypeMarket, buyEvent, sellEvent)
	if err != nil {
		return err
	}

	_ = pairProcessor.ProcessBuyOrder()

	if pair.GetStage() == pairs_types.InputIntoPositionStage {
		collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.OutputOfPositionStage)
		config.Save()
	}
	if pair.GetStage() == pairs_types.OutputOfPositionStage {
		orderExecutionGuard := pairProcessor.ProcessSellTakeProfitOrder()
		<-orderExecutionGuard
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
	account *spot_account.Account,
	stopEvent chan os.Signal,
	updateTime time.Duration,
	minuteOrderLimit *exchange_types.RateLimits,
	dayOrderLimit *exchange_types.RateLimits,
	minuteRawRequestLimit *exchange_types.RateLimits,
	orderStatusEvent chan *binance.WsUserDataEvent) (err error) {
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
			account,
			stopEvent,
			updateTime,
			minuteOrderLimit,
			dayOrderLimit,
			minuteRawRequestLimit,
			orderStatusEvent)

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		RunSpotScalping(
			config,
			client,
			degree,
			limit,
			pair,
			account,
			stopEvent,
			updateTime,
			minuteOrderLimit,
			dayOrderLimit,
			minuteRawRequestLimit,
			orderStatusEvent)

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {

		RunSpotTrading(
			config,
			client,
			degree,
			limit,
			pair,
			account,
			stopEvent,
			updateTime,
			minuteOrderLimit,
			dayOrderLimit,
			minuteRawRequestLimit,
			orderStatusEvent)

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
	}
	return nil
}
