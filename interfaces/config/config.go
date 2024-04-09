package config

import pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"

type (
	Pairs interface {
		GetAccountType() pairs_types.AccountType
		GetStrategy() pairs_types.StrategyType
		GetStage() pairs_types.StageType
		SetStage(stage pairs_types.StageType)
		GetPair() string
		GetTargetSymbol() string
		GetBaseSymbol() string
		GetLimitInputIntoPosition() float64
		GetLimitInPosition() float64
		GetLimitOnTransaction() float64
		GetBuyDelta() float64
		GetSellDelta() float64
		GetBuyQuantity() float64
		GetSellQuantity() float64
		GetBuyValue() float64
		GetSellValue() float64
		SetBuyQuantity(quantity float64)
		SetSellQuantity(quantity float64)
		SetBuyValue(value float64)
		SetSellValue(value float64)
		GetMiddlePrice() float64
		GetProfit(currentPrice float64) float64
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
		Lock()
		Unlock()
		GetConfigurations() Configuration
	}
)
