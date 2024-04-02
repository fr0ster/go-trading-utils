package config

import pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"

type (
	Pairs interface {
		GetAccountType() pairs_types.AccountType
		GetPair() string
		GetTargetSymbol() string
		GetBaseSymbol() string
		GetLimit() float64
		GetDelta() float64
		GetQuantity() float64
		GetValue() float64
		SetLimit(limit float64)
		SetQuantity(quantity float64)
		SetValue(value float64)
	}
	Configuration interface {
		GetAPIKey() string
		GetSecretKey() string
		GetUseTestNet() bool
		GetPair(pair string) Pairs
		GetPairs(account_type ...pairs_types.AccountType) (*[]Pairs, error)
		SetPairs([]Pairs) error
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		GetConfigurations() Configuration
	}
)
