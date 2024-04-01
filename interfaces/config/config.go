package config

type (
	Pairs interface {
		GetPair() string
		GetTargetSymbol() string
		GetBaseSymbol() string
		GetLimit() float64
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
		GetPairs() ([]Pairs, error)
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		GetConfigurations() Configuration
	}
)
