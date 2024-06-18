package config

import (
	connection_interfaces "github.com/fr0ster/go-trading-utils/interfaces/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	"github.com/sirupsen/logrus"
)

type (
	Configuration interface {
		GetConnection() connection_interfaces.Connection

		GetPair(
			account pairs_types.AccountType,
			strategy pairs_types.StrategyType,
			stage pairs_types.StageType,
			pair string) *pairs_types.Pairs
		SetPair(*pairs_types.Pairs)

		GetPairs(account_type ...pairs_types.AccountType) ([]*pairs_types.Pairs, error)
		SetPairs([]*pairs_types.Pairs) error

		GetLogLevel() logrus.Level
		SetLogLevel(level logrus.Level)

		GetReloadConfig() bool
		SetReloadConfig(bool)

		GetObservePriceLiquidation() bool
		SetObservePriceLiquidation(bool)
		GetObservePositionLoss() bool
		SetObservePositionLoss(bool)
		GetClosePositionOnRestart() bool
		SetClosePositionOnRestart(bool)

		GetBalancingOfMargin() bool
		SetBalancingOfMargin(bool)

		GetPercentsToStopSettingNewOrder() float64
		SetPercentsToStopSettingNewOrder(float64)
		GetPercentToDecreasePosition() float64
		SetPercentToDecreasePosition(float64)

		GetObserverTimeOutMillisecond() int
		SetObserverTimeOutMillisecond(int)

		GetUsingBreakEvenPrice() bool
		SetUsingBreakEvenPrice(bool)

		GetBuyDeltaLoss() float64
		SetBuyDeltaLoss(float64)
		GetSellDeltaLoss() float64
		SetSellDeltaLoss(float64)

		GetDeltaStepPercent() float64
		SetDeltaStepPercent(float64)

		GetClosePositionByTakeProfitMarketOrder() bool
		SetClosePositionByTakeProfitMarketOrder(bool)

		GetDynamicDelta() bool
		SetDynamicDelta(bool)

		GetDynamicQuantity() bool
		SetDynamicQuantity(bool)
	}
	ConfigurationFile interface {
		GetFileName() string
		Save() error
		Load() error
		Lock()
		Unlock()
		GetConfigurations() Configuration
	}
)
