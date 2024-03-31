package config

type (
	Pairs interface {
		GetPair() string
		GetTargetSymbol() string
		GetBaseSymbol() string
		GetLimit() float64
		GetQuantity() float64
		GetValue() float64
	}
	Configuration interface {
		GetAPIKey() string
		GetSecretKey() string
		GetUseTestNet() bool
		GetPairs(pair string) Pairs
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		GetConfigurations() Configuration
	}
)
