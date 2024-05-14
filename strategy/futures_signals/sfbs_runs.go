package futures_signals

import (
	"fmt"
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
	interval  = "1m"
)

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair pairs_interfaces.Pairs,
	stopEvent chan os.Signal,
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

	account, err := futures_account.New(client, degree, []string{pair.GetBaseSymbol()}, []string{pair.GetTargetSymbol()})
	if err != nil {
		return
	}

	PairInit(client, config, account, pair)

	pairBookTickerObserver, _ := NewPairBookTickerObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	pairBookTickerObserver.StartUpdateGuard()
	pairObserver, _ := NewPairObserver(client, pair, degree, limit, deltaUp, deltaDown, stopEvent)
	riskEvent := pairObserver.StartRiskSignal()
	askUp, askDown, bidUp, bidDown := pairBookTickerObserver.StartPriceChangesSignal()
	buyEvent, sellEvent := pairBookTickerObserver.StartBuyOrSellSignal()

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
			config, client, pair, futures.OrderTypeMarket, debug)
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
		if pair.GetStage() == pairs_types.PositionClosedStage {
			return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
		}
		collectionOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)

		_ = pairProcessor.ProcessBuyOrder(buyEvent)

		<-collectionOutEvent
		pair.SetStage(pairs_types.PositionClosedStage)
		config.Save()
		stopEvent <- os.Interrupt
		return nil

		// Відпрацьовуємо Trading стратегію
	} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
		if pair.GetStage() == pairs_types.PositionClosedStage {
			return fmt.Errorf("pair %v has wrong stage %v", pair.GetPair(), pair.GetStage())
		}
		if pair.GetStage() == pairs_types.InputIntoPositionStage || pair.GetStage() == pairs_types.WorkInPositionStage {
			_ = pairProcessor.ProcessBuyOrder(buyEvent)   // Запускаємо процес купівлі
			_ = pairProcessor.ProcessSellOrder(sellEvent) // Запускаємо процес продажу
		}
		if pair.GetStage() == pairs_types.InputIntoPositionStage {
			collectionOutEvent := pairObserver.StartWorkInPositionSignal(triggerEvent)
			<-collectionOutEvent
			pair.SetStage(pairs_types.WorkInPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.WorkInPositionStage {
			workingOutEvent := pairObserver.StopWorkInPositionSignal(triggerEvent)
			<-workingOutEvent
			pair.SetStage(pairs_types.OutputOfPositionStage)
			config.Save()
		}
		if pair.GetStage() == pairs_types.OutputOfPositionStage {
			positionClosed := pairObserver.ClosePositionSignal(triggerEvent)
			<-positionClosed
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
