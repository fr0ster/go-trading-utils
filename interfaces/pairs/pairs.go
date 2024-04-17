package pairs

import (
	"time"

	"github.com/adshao/go-binance/v2"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	Pairs interface {
		GetConnection() *connection_types.Connection
		SetConnection(connection *connection_types.Connection)
		GetInitialBalance() float64
		SetInitialBalance(balance float64)
		GetAccountType() pairs_types.AccountType
		GetStrategy() pairs_types.StrategyType
		GetStage() pairs_types.StageType
		SetStage(stage pairs_types.StageType)
		GetPair() string
		GetTargetSymbol() string
		GetBaseSymbol() string
		GetSleepingTime() time.Duration
		GetTakingPositionSleepingTime() time.Duration
		GetLimitInputIntoPosition() float64
		GetLimitOutputOfPosition() float64
		GetLimitOnPosition() float64
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
		GetBuyCommission() float64
		SetBuyCommission(commission float64)
		GetSellCommission() float64
		SetSellCommission(commission float64)
		AddCommission(commission *binance.Fill)
		GetCommission() pairs_types.Commission
		SetCommission(commission pairs_types.Commission)
		CalcMiddlePrice() error
		GetMiddlePrice() float64
		SetMiddlePrice(price float64)
		GetProfit(currentPrice float64) float64
		CheckingPair() bool
	}
)
