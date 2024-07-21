package spot_signals

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2"

	grid "github.com/fr0ster/go-trading-utils/strategy/spot_signals/grid"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

func RunSpotHolding(
	client *binance.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	stopEvent chan struct{},
	updateTime time.Duration,
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
	updateTime time.Duration,
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
	updateTime time.Duration,
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
	updateTime time.Duration,
	debug bool,
	wg *sync.WaitGroup,
	depths ...*depth_types.Depths) {
	// var (
	// 	depth *depth_types.Depths
	// )
	// if len(depths) > 0 {
	// 	depth = depths[0]
	// } else {
	// 	depth = depth_types.New(
	// 		degree,
	// 		pair.GetPair(),
	// 		1000*time.Millisecond,
	// 		spot_depth.GetterStartPartialDepthStreamCreator(
	// 			depth_types.DepthStreamLevel5,
	// 			depth_types.DepthStreamRate100ms,
	// 			spot_depth.GetterPartialDepthEventCallBackCreator(),
	// 			spot_depth.GetterWsErrorHandlerCreator()),
	// 		spot_depth.GetterInitCreator(depth_types.DepthAPILimit20, client))
	// }
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
					updateTime,
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
					updateTime,
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
					updateTime,
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
					pair.GetPercentToTarget(),    // targetPercent
					pair.GetCallbackRate(),       // callbackRate
					stopEvent,                    // stopEvent
					wg))                          // wg

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			logrus.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
	}()
}
