package spot_signals

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	grid "github.com/fr0ster/go-trading-utils/strategy/spot_signals/grid"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
)

func RunSpotHolding(
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

func RunSpotScalping(
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

func RunSpotTrading(
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return nil
}

func Run(
	config *config_types.ConfigFile,
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	debug bool,
	wg *sync.WaitGroup,
	depths ...*depth_types.Depths) {
	wg.Add(1)
	go func() {
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			logrus.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			logrus.Error(
				RunSpotHolding(
					client,
					degree,
					limit,
					pair,
					stopEvent,
					wg))

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			logrus.Error(
				RunSpotScalping(
					client,
					degree,
					limit,
					pair,
					stopEvent,
					wg))

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			logrus.Error(
				RunSpotTrading(
					client,
					degree,
					limit,
					pair,
					stopEvent,
					wg))

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			logrus.Error(
				grid.RunSpotGridTrading(
					client,
					pair.GetPair(),               // symbol
					pair.GetLimitOnPosition(),    // limitOnPosition
					pair.GetLimitOnTransaction(), // limitOnTransaction
					pair.GetUpBound(),            // upBound
					pair.GetLowBound(),           // lowBound
					pair.GetDeltaPrice(),         // deltaPrice
					pair.GetDeltaQuantity(),      // deltaQuantity
					pair.GetMinSteps(),           // minSteps
					pair.GetCallbackRate(),       // callbackRate
					stopEvent,                    // stopEvent
					wg))                          // wg

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			logrus.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
	}()
}
