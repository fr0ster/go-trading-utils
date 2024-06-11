package pairs

import (
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	Pairs interface {
		GetInitialBalance() float64
		SetInitialBalance(float64)
		GetCurrentBalance() float64
		SetCurrentBalance(float64)

		GetInitialPositionBalance() float64
		SetInitialPositionBalance(float64)

		GetAccountType() pairs_types.AccountType

		GetStrategy() pairs_types.StrategyType
		SetStrategy(pairs_types.StrategyType)
		GetStage() pairs_types.StageType
		SetStage(pairs_types.StageType)

		GetPair() string

		GetTargetSymbol() string
		GetBaseSymbol() string

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

		GetDeltaStep() float64

		GetBuyDelta() float64
		GetSellDelta() float64

		GetBuyQuantity() float64
		GetSellQuantity() float64

		GetBuyValue() float64
		GetSellValue() float64

		SetBuyQuantity(float64)
		SetSellQuantity(float64)

		SetDeltaStep(float64)

		SetBuyValue(float64)
		SetSellValue(float64)

		GetBuyCommission() float64
		SetBuyCommission(float64)

		GetSellCommission() float64
		SetSellCommission(float64)

		SetBuyData(float64, float64, float64)
		SetSellData(float64, float64, float64)

		CalcMiddlePrice() error
		GetMiddlePrice() float64
		SetMiddlePrice(float64)

		GetProfit(float64) float64

		CheckingPair() bool
	}
)
