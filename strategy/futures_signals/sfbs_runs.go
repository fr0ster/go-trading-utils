package futures_signals

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	futures_depth "github.com/fr0ster/go-trading-utils/binance/futures/markets/depth"

	grid "github.com/fr0ster/go-trading-utils/strategy/futures_signals/grid"
	trading "github.com/fr0ster/go-trading-utils/strategy/futures_signals/trading"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	depth_types "github.com/fr0ster/go-trading-utils/types/depth"
	depths_types "github.com/fr0ster/go-trading-utils/types/depth/depths"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	quit chan struct{},
	debug bool,
	wg *sync.WaitGroup,
	depths ...*depth_types.Depths) {
	var (
		depth *depth_types.Depths
	)
	if len(depths) > 0 {
		depth = depths[0]
	} else {
		depth = depth_types.New(
			quit,
			degree,
			pair.GetPair(),
			1000*time.Millisecond,
			futures_depth.GetStartDepthStream(
				depth,
				depths_types.DepthStreamLevel5,
				depths_types.DepthStreamRate100ms,
				futures_depth.GetDepthEventCallBack(depth),
				futures_depth.GetWsErrorHandler(depth)),
			func(d *depth_types.Depths) error { return futures_depth.Init(d, depths_types.DepthAPILimit20, client) })
	}
	wg.Add(1)
	go func() {
		var err error
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			err = fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			err = fmt.Errorf("holding strategy shouldn't be implemented for futures")

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			err = grid.RunFuturesGridTradingV1(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetPercentToTarget(),                                     // targetPercent
				pair.GetDepthsN(),                                             // limitDepth
				2,                                                             // expBase
				pair.GetCallbackRate(),                                        // callbackRate
				pair.GetProgression(),                                         // progression
				quit,                                                          // quit
				wg)                                                            // wg

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			err = trading.RunFuturesTrading(
				client,                       // client
				pair.GetPair(),               // pair
				degree,                       // degree
				limit,                        // limit
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				pair.GetPercentToTarget(),    // targetPercent
				pair.GetDepthsN(),            // limitDepth
				2,                            // expBase
				pair.GetCallbackRate(),       // callbackRate
				futures.SideTypeBuy,          // upOrderSideOpen
				futures.OrderTypeStop,        // upPositionNewOrderType
				futures.SideTypeSell,         // downOrderSideOpen
				futures.OrderTypeStop,        // downPositionNewOrderType
				futures.OrderTypeTakeProfit,  // shortPositionTPOrderType
				futures.OrderTypeStop,        // shortPositionSLOrderType
				futures.OrderTypeTakeProfit,  // longPositionTPOrderType
				futures.OrderTypeStop,        // longPositionSLOrderType
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg)

			// Відпрацьовуємо Grid стратегію
		} else if pair.GetStrategy() == pairs_types.GridStrategyType {
			err = grid.RunFuturesGridTradingV1(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetPercentToTarget(),                                     // targetPercent
				pair.GetDepthsN(),                                             // limitPercent
				2,                                                             // expBase
				pair.GetCallbackRate(),                                        // callbackRate
				pair.GetProgression(),                                         // progression
				quit,                                                          // quit
				wg)                                                            // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV2 {
			err = grid.RunFuturesGridTradingV2(
				client,                       // client
				pair,                         // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				10,                           // targetPercent
				75,                           // limitPercent
				2,                            // expBase
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				quit,                  // quit
				pair.GetProgression(), // progression
				wg)                    // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV3 {
			// Відкриваємо позицію лімітними ордерами,
			// Збільшуємо та зменшуємо позицію трейлінг стопами
			// відкриваємо ордера на продаж та купівлю з однаковою кількістью
			// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
			err = grid.RunFuturesGridTradingV3(
				client,                       // client
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				pair.GetPercentToTarget(),    // targetPercent
				pair.GetDepthsN(),            // limitDepth
				2,                            // expBase
				pair.GetCallbackRate(),       // callbackRate
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg)                           // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV4 {
			// Відкриваємо позицію лімітними ордерами,
			// Збільшуємо та зменшуємо позицію лімітними ордерами
			// відкриваємо ордера на продаж та купівлю з однаковою кількістью
			// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
			err = grid.RunFuturesGridTradingV4(
				client,                       // client
				degree,                       // degree
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				pair.GetPercentToTarget(),    // targetPercent
				pair.GetDepthsN(),            // limitDepth
				2,                            // expBase
				pair.GetCallbackRate(),       // callbackRate
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg,                           // wg
				depth)

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV5 {
			// Відкриваємо позицію лімітними ордерами,
			// Збільшуємо та зменшуємо позицію тейк профіт ордерами
			// відкриваємо ордера на продаж та купівлю з однаковою кількістью
			// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
			err = grid.RunFuturesGridTradingV5(
				client,                       // client
				degree,                       // degree
				pair.GetPair(),               // pair
				pair.GetLimitOnPosition(),    // limitOnPosition
				pair.GetLimitOnTransaction(), // limitOnTransaction
				pair.GetUpBound(),            // upBound
				pair.GetLowBound(),           // lowBound
				pair.GetDeltaPrice(),         // deltaPrice
				pair.GetDeltaQuantity(),      // deltaQuantity
				pair.GetMarginType(),         // marginType
				pair.GetLeverage(),           // leverage
				pair.GetMinSteps(),           // minSteps
				pair.GetPercentToTarget(),    // targetPercent
				pair.GetDepthsN(),            // limitDepth
				2,                            // expBase
				pair.GetCallbackRate(),       // callbackRate
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg,                           // wg
				depth)

			// Невідома стратегія, виводимо попередження та завершуємо програму
		} else {
			err = fmt.Errorf("unknown strategy: %v", pair.GetStrategy())
		}
		if err != nil {
			logrus.Error(err)
			close(quit)
		}
	}()
}
