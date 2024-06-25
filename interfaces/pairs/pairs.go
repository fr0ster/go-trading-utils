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

		GetUnRealizedProfitLowBound() float64
		GetUnRealizedProfitUpBound() float64

		GetUpBound() float64
		GetLowBound() float64

		GetDeltaPrice() float64
		SetDeltaPrice(float64)

		GetBuyQuantity() float64
		GetSellQuantity() float64

		GetIsArithmetic() bool
		SetIsArithmetic(bool)

		GetBuyValue() float64
		GetSellValue() float64

		SetBuyQuantity(float64)
		SetSellQuantity(float64)

		SetDeltaStepPerMille(float64)

		SetBuyValue(float64)
		SetSellValue(float64)

		SetBuyData(float64, float64, float64)
		SetSellData(float64, float64, float64)

		GetDeltaQuantity() float64
		SetDeltaQuantity(float64)

		GetCallbackRate() float64
		SetCallbackRate(float64)

		GetMiddlePrice() float64

		GetProfit(float64) float64

		CheckingPair() bool
	}
)
