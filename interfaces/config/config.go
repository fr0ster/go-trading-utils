package config

import (
	connection_interfaces "github.com/fr0ster/go-trading-utils/interfaces/connection"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	"github.com/sirupsen/logrus"
)

type (
	Configuration interface {
		GetConnection() connection_interfaces.Connection
		GetPair(pair string) pairs_interfaces.Pairs
		SetPair(pairs_interfaces.Pairs)
		GetPairs(account_type ...pairs_types.AccountType) (*[]pairs_interfaces.Pairs, error)
		SetPairs([]pairs_interfaces.Pairs) error
		GetLogLevel() logrus.Level
		SetLogLevel(level logrus.Level)
		GetReloadConfig() bool
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		Lock()
		Unlock()
		GetConfigurations() Configuration
	}
)
