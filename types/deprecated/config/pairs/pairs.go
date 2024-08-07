package pairs

import (
	"strings"

	"github.com/fr0ster/go-trading-utils/types"
	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	depth_types "github.com/fr0ster/go-trading-utils/types/depths"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"

	"github.com/google/btree"
)

type (
	Pairs struct {
		AccountType  types.AccountType  `json:"account_type"`  // Тип акаунта
		StrategyType types.StrategyType `json:"strategy_type"` // Тип стратегії
		StageType    types.StageType    `json:"stage_type"`    // Cтадія стратегії

		Pair string `json:"symbol"` // Пара

		// Для USDT_FUTURE/COIN_FUTURE
		MarginType types.MarginType `json:"margin_type"` // Тип маржі
		Leverage   int              `json:"leverage"`    // Маржинальне плече

		LimitOnPosition    items_types.ValueType        `json:"limit_on_position"`    // Ліміт на позицію, відсоток від балансу базової валюти
		LimitOnTransaction items_types.ValuePercentType `json:"limit_on_transaction"` // Ліміт на транзакцію, відсоток від ліміту на позицію

		UpAndLowBound items_types.PricePercentType `json:"up_and_low_bound"` // Верхня та Нижня межи ціни, най буде відсоток від ціни безубитку позиції
		MinSteps      int                          `json:"min_steps"`        // Мінімальна кількість кроків

		DeltaPrice    items_types.PricePercentType    `json:"delta_price"`    // Дельта для купівлі/продажу
		DeltaQuantity items_types.QuantityPercentType `json:"delta_quantity"` // Кількість для купівлі/продажу
		Progression   types.ProgressionType           `json:"progression"`    // Тип прогресії

		Value items_types.ValueType `json:"value"` // Вартість позиції

		CallbackRate items_types.PricePercentType `json:"callback_rate"` // callbackRate для TRAILING_STOP_MARKET

		DepthsN int `json:"depths_n"` // Глибина стакана
	}
)

// Less implements btree.Item.
func (pr *Pairs) Less(item btree.Item) bool {
	other := item.(*Pairs)
	if pr.AccountType != other.AccountType && pr.AccountType != "" && other.AccountType != "" {
		return pr.AccountType < other.AccountType
	}
	if pr.StrategyType != other.StrategyType && pr.StrategyType != "" && other.StrategyType != "" {
		return pr.StrategyType < other.StrategyType
	}
	if pr.StageType != other.StageType && pr.StageType != "" && other.StageType != "" {
		return pr.StageType < other.StageType
	}
	return pr.Pair < other.Pair
}

// Equals implements btree.Item.
func (pr *Pairs) Equals(item btree.Item) bool {
	other := item.(*Pairs)
	return pr.AccountType == other.AccountType &&
		pr.StrategyType == other.StrategyType &&
		pr.StageType == other.StageType &&
		pr.Pair == other.Pair
}

// Get types.AccountType implements Pairs.
func (pr *Pairs) GetAccountType() types.AccountType {
	return pr.AccountType
}

// GetStrategy implements Pairs.
func (pr *Pairs) GetStrategy() types.StrategyType {
	return pr.StrategyType
}

// SetStrategy implements Pairs.
func (pr *Pairs) SetStrategy(strategy types.StrategyType) {
	pr.StrategyType = strategy
}

// GetStage implements Pairs.
func (pr *Pairs) GetStage() types.StageType {
	return pr.StageType
}

// SetStage implements Pairs.
func (pr *Pairs) SetStage(stage types.StageType) {
	pr.StageType = stage
}

// GetSymbol implements Pairs.
func (pr *Pairs) GetPair() string {
	return pr.Pair
}

// GetMarginType implements Pairs.
func (pr *Pairs) GetMarginType() types.MarginType {
	return pr.MarginType
}

// SetMarginType implements pairs.Pairs.
func (pr *Pairs) SetMarginType(marginType types.MarginType) {
	pr.MarginType = types.MarginType(strings.ToUpper(string(marginType)))
}

// SetMarginType implements Pairs.
func (pr *Pairs) GetLeverage() int {
	return pr.Leverage
}

// SetLeverage implements Pairs.
func (pr *Pairs) SetLeverage(leverage int) {
	pr.Leverage = leverage
}

func (pr *Pairs) GetLimitOnPosition() items_types.ValueType {
	return pr.LimitOnPosition
}

func (pr *Pairs) GetLimitOnTransaction() items_types.ValuePercentType {
	return pr.LimitOnTransaction
}

func (pr *Pairs) GetUpAndLowBound() items_types.PricePercentType {
	return pr.UpAndLowBound
}

func (pr *Pairs) GetMinSteps() int {
	return pr.MinSteps
}

func (pr *Pairs) GetDeltaPrice() items_types.PricePercentType {
	return pr.DeltaPrice
}

func (pr *Pairs) GetDeltaQuantity() items_types.QuantityPercentType {
	return pr.DeltaQuantity
}

func (pr *Pairs) GetProgression() types.ProgressionType {
	return pr.Progression
}

func (pr *Pairs) GetValue() items_types.ValueType {
	return pr.Value
}

func (pr *Pairs) SetLimitOnPosition(val items_types.ValueType) {
	pr.LimitOnPosition = val
}

func (pr *Pairs) SetLimitOnTransaction(val items_types.ValuePercentType) {
	pr.LimitOnTransaction = val
}

func (pr *Pairs) SetUpAndLowBound(val items_types.PricePercentType) {
	pr.UpAndLowBound = val
}

func (pr *Pairs) SetMinSteps(val int) {
	pr.MinSteps = val
}

func (pr *Pairs) SetDeltaPrice(val items_types.PricePercentType) {
	pr.DeltaPrice = val
}

func (pr *Pairs) SetDeltaQuantity(quantity items_types.QuantityPercentType) {
	pr.DeltaQuantity = quantity
}

func (pr *Pairs) SetProgression(val types.ProgressionType) {
	pr.Progression = val
}

func (pr *Pairs) SetValue(value items_types.ValueType) {
	pr.Value = value
}

func (pr *Pairs) GetCallbackRate() items_types.PricePercentType {
	return pr.CallbackRate
}

func (pr *Pairs) SetCallbackRate(rate items_types.PricePercentType) {
	pr.CallbackRate = rate
}

func (pr *Pairs) GetDepthsN() depth_types.DepthAPILimit {
	if pr.DepthsN == 0 {
		return depth_types.DepthAPILimit(50)
	} else {
		return depth_types.DepthAPILimit(pr.DepthsN)
	}
}

func (pr *Pairs) SetDepthsN(n depth_types.DepthAPILimit) {
	pr.DepthsN = int(n)
}

func New(
	connection *connection_types.Connection,
	AccountType types.AccountType,
	StrategyType types.StrategyType,
	stageType types.StageType,
	pair string,
	targetSymbol string,
	baseSymbol string,
) *Pairs {
	return &Pairs{
		AccountType:        AccountType,
		StrategyType:       StrategyType,
		StageType:          stageType,
		Pair:               pair,
		LimitOnPosition:    100.0, // 100$
		LimitOnTransaction: 1.0,   // 1%
		DeltaPrice:         1.0,   // 1%
		DeltaQuantity:      10.0,  // 10%
		Progression:        "GEOMETRIC",
		Value:              0.0,
		CallbackRate:       0.1, // 0.1%
		DepthsN:            50,  // 50
	}
}
