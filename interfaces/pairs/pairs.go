package pairs

import (
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	Pairs interface {
		GetAccountType() pairs_types.AccountType

		GetStrategy() pairs_types.StrategyType
		SetStrategy(pairs_types.StrategyType)
		GetStage() pairs_types.StageType
		SetStage(pairs_types.StageType)

		GetPair() string

		GetMarginType() pairs_types.MarginType
		SetMarginType(pairs_types.MarginType)
		GetLeverage() int
		SetLeverage(int)

		GetLimitInputIntoPosition() float64
		GetLimitOutputOfPosition() float64

		GetLimitOnPosition() float64
		GetLimitOnTransaction() float64

		GetUpBound() float64
		GetLowBound() float64

		GetDeltaPrice() float64
		SetDeltaPrice(float64)

		GetDeltaQuantity() float64
		SetDeltaQuantity(float64)

		GetMinSteps() int

		GetProgression() pairs_types.ProgressionType
		SetProgression(pairs_types.ProgressionType)

		GetValue() float64
		SetValue(float64)

		GetCallbackRate() float64
		SetCallbackRate(float64)
	}
)
