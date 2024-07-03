package futures_signals

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/adshao/go-binance/v2/futures"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

func RunFuturesHolding(wg *sync.WaitGroup) (err error) {
	defer wg.Done()
	return fmt.Errorf("it should be implemented for futures")
}

func RunScalpingHolding(
	client *futures.Client,
	pair string,
	limitOnPosition float64,
	limitOnTransaction float64,
	upBound float64,
	lowBound float64,
	deltaPrice float64,
	deltaQuantity float64,
	marginType pairs_types.MarginType,
	leverage int,
	minSteps int,
	callbackRate float64,
	percentsToStopSettingNewOrder float64,
	progression pairs_types.ProgressionType,
	quit chan struct{},
	wg *sync.WaitGroup) (err error) {
	return RunFuturesGridTradingV1(
		client,
		pair,
		limitOnPosition,
		limitOnTransaction,
		upBound,
		lowBound,
		deltaPrice,
		deltaQuantity,
		marginType,
		leverage,
		minSteps,
		percentsToStopSettingNewOrder,
		callbackRate,
		progression,
		quit,
		wg)
}

func Run(
	config *config_types.ConfigFile,
	client *futures.Client,
	degree int,
	limit int,
	pair *pairs_types.Pairs,
	quit chan struct{},
	debug bool,
	wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		var err error
		// Відпрацьовуємо Arbitrage стратегію
		if pair.GetStrategy() == pairs_types.ArbitrageStrategyType {
			err = fmt.Errorf("arbitrage strategy is not implemented yet for %v", pair.GetPair())

			// Відпрацьовуємо  Holding стратегію
		} else if pair.GetStrategy() == pairs_types.HoldingStrategyType {
			err = RunFuturesHolding(wg)

			// Відпрацьовуємо Scalping стратегію
		} else if pair.GetStrategy() == pairs_types.ScalpingStrategyType {
			err = RunScalpingHolding(
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
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetProgression(), // progression
				quit,                  // quit
				wg)                    // wg

			// Відпрацьовуємо Trading стратегію
		} else if pair.GetStrategy() == pairs_types.TradingStrategyType {
			err = RunFuturesTrading(
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
			err = RunFuturesGridTradingV1(
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
				pair.GetCallbackRate(),       // callbackRate
				config.GetConfigurations().GetPercentsToStopSettingNewOrder(), // percentsToStopSettingNewOrder
				pair.GetProgression(), // progression
				quit,                  // quit
				wg)                    // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV2 {
			err = RunFuturesGridTradingV2(
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
			err = RunFuturesGridTradingV3(
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
				pair.GetCallbackRate(),       // callbackRate
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg)                           // wg

		} else if pair.GetStrategy() == pairs_types.GridStrategyTypeV4 {
			// Відкриваємо позицію лімітними ордерами,
			// Збільшуємо та зменшуємо позицію лімітними ордерами
			// відкриваємо ордера на продаж та купівлю з однаковою кількістью
			// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
			err = RunFuturesGridTradingV4(
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
				pair.GetCallbackRate(),       // callbackRate
				pair.GetProgression(),        // progression
				quit,                         // quit
				wg)                           // wg

			// } else if pair.GetStrategy() == pairs_types.GridStrategyTypeV5 {
			// 	// Відкриваємо позицію лімітними ордерами,
			// 	// Збільшуємо та зменшуємо позицію тейк профіт ордерами
			// 	// відкриваємо ордера на продаж та купівлю з однаковою кількістью
			// 	// Ціну визначаємо або дінамічно і кожний новий ордер який збільшує позицію
			// 	err = RunFuturesGridTradingV3(
			// 		client,                              // client
			// 		pair.GetPair(),                      // pair
			// 		pair.GetLimitOnPosition(),           // limitOnPosition
			// 		pair.GetLimitOnTransaction(),        // limitOnTransaction
			// 		pair.GetUpBound(),                   // upBound
			// 		pair.GetLowBound(),                  // lowBound
			// 		pair.GetDeltaPrice(),                // deltaPrice
			// 		pair.GetDeltaQuantity(),             // deltaQuantity
			// 		pair.GetMarginType(),                // marginType
			// 		pair.GetLeverage(),                  // leverage
			// 		pair.GetMinSteps(),                  // minSteps
			// 		pair.GetCallbackRate(),              // callbackRate
			// 		futures.SideTypeSell,                // upOrderSideOpen
			// 		futures.OrderTypeLimit,              // upPositionNewOrderType
			// 		futures.SideTypeBuy,                 // downOrderSideOpen
			// 		futures.OrderTypeLimit,              // downPositionNewOrderType
			// 		futures.OrderTypeTrailingStopMarket, // shortPositionIncOrderType
			// 		futures.OrderTypeTrailingStopMarket, // shortPositionDecOrderType
			// 		futures.OrderTypeTrailingStopMarket, // longPositionIncOrderType
			// 		futures.OrderTypeTrailingStopMarket, // longPositionDecOrderType
			// 		pair.GetProgression(),               // progression
			// 		quit,                                // quit
			// 		wg)                                  // wg

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
