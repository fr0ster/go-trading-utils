package config

type (
	Configuration interface {
		GetSymbol() string
		GetLimit() float64
		GetQuantity() float64
		GetValue() float64
		GetAPIKey() string
		GetSecretKey() string
		GetUseTestNet() bool
	}
	ConfigurationFile interface {
		Save() error
		Load() error
		GetConfigurations() Configuration
	}
)
