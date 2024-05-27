package pairs

import (
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	Pairs interface {
		GetInitialBalance() float64
		SetInitialBalance(balance float64)
		GetCurrentBalance() float64
		SetCurrentBalance(balance float64)

		GetInitialPositionBalance() float64
		SetInitialPositionBalance(balance float64)

		GetAccountType() pairs_types.AccountType

		GetStrategy() pairs_types.StrategyType
		SetStrategy(strategy pairs_types.StrategyType)
		GetStage() pairs_types.StageType
		SetStage(stage pairs_types.StageType)

		GetPair() string

		GetTargetSymbol() string
		GetBaseSymbol() string

		GetMarginType() pairs_types.MarginType
		SetMarginType(marginType pairs_types.MarginType)
		GetLeverage() int
		SetLeverage(leverage int)

		GetLimitInputIntoPosition() float64
		GetLimitOutputOfPosition() float64

		GetLimitOnPosition() float64
		GetLimitOnTransaction() float64

		GetUpBound() float64
		GetLowBound() float64

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

		GetBuyCommission() float64
		SetBuyCommission(commission float64)

		GetSellCommission() float64
		SetSellCommission(commission float64)

		SetBuyData(quantity, value, commission float64)
		SetSellData(quantity, value, commission float64)

		SetBuyDelta(delta float64)
		SetSellDelta(delta float64)

		CalcMiddlePrice() error
		GetMiddlePrice() float64
		SetMiddlePrice(price float64)

		GetProfit(currentPrice float64) float64

		CheckingPair() bool
	}
)
