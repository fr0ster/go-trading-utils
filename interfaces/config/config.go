package config

import (
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	Connection interface {
		GetAPIKey() string
		SetApiKey(key string)
		GetSecretKey() string
		SetSecretKey(key string)
		GetUseTestNet() bool
		SetUseTestNet(useTestNet bool)
	}
	Configuration interface {
		GetSpotConnection() Connection
		GetFuturesConnection() Connection
		GetPair(pair string) pairs_interfaces.Pairs
		GetPairs(account_type ...pairs_types.AccountType) (*[]pairs_interfaces.Pairs, error)
		SetPairs([]pairs_interfaces.Pairs) error
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		Lock()
		Unlock()
		GetConfigurations() Configuration
	}
)
