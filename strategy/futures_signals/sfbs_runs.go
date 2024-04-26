package futures_signals

import (
	"fmt"
	_ "net/http/pprof"

	"os"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"

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

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal) (err error) {

	account, err := futures_account.New(client, degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	PairInit(client, config, account, pair)

	pairObserver := NewPairObserver(client, account, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	pairObserver.StartBookTickersUpdateGuard()
	riskEvent := pairObserver.StartRiskSignal()
	askUp, askDown, bidUp, bidDown := pairObserver.StartPriceSignal()

	triggerEvent := make(chan bool)
	go func() {
		for {
			select {
			case <-stopEvent:
				stopEvent <- os.Interrupt
				return
			case <-riskEvent:
				stopEvent <- os.Interrupt
				return
			case <-askUp:
				triggerEvent <- true
			case <-askDown:
				triggerEvent <- true
			case <-bidUp:
				triggerEvent <- true
			case <-bidDown:
				triggerEvent <- true
			}
		}
	}()

	pairProcessor, err :=
		NewPairProcessor(
			config, client, pair, futures.OrderTypeMarket, nil, nil, askUp, askDown, bidUp, bidDown)
	if err != nil {
		return err
	}

	// Відпрацьовуємо Arbitrage стратегію
	if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
		logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

		// Відпрацьовуємо  Holding стратегію
	} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			logrus.Warnf("Uncorrected strategy: %v", pair.GetStrategy())
			stopEvent <- os.Interrupt
			return fmt.Errorf("holding strategy should not be used for %v", pair.GetPair())
		}

		// Відпрацьовуємо Scalping стратегію
	} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
		collectionOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)

		_ = pairProcessor.ProcessBuyOrder()

		<-collectionOutEvent
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
		return nil

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
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

		// Невідома стратегія, виводимо попередження та завершуємо програму
	} else {
		logrus.Warnf("Unknown strategy: %v", pair.GetStrategy())
		stopEvent <- os.Interrupt
		return
	}
	return nil
}
