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

		GetObservePriceLiquidation() bool
		GetObservePositionLoss() bool
		GetClosePositionOnRestart() bool

		GetBalancingOfMargin() bool

		GetPercentsToStopSettingNewOrder() float64
		GetPercentToDecreasePosition() float64

		GetObserverTimeOutMillisecond() int

		GetUsingBreakEvenPrice() bool

		GetDynamicDelta() bool

		GetBuyDeltaLoss() float64
		GetSellDeltaLoss() float64
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
