package pairs

import (
	"strings"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"

	"github.com/google/btree"
)

const (
	// SpotAccountType is a constant for spot account type.
	// SPOT/MARGIN/ISOLATED_MARGIN/USDT_FUTURE/COIN_FUTURE
	SpotAccountType           AccountType = "SPOT"
	MarginAccountType         AccountType = "MARGIN"
	IsolatedMarginAccountType AccountType = "ISOLATED_MARGIN"
	USDTFutureType            AccountType = "USDT_FUTURE"
	CoinFutureType            AccountType = "COIN_FUTURE"
	// SpotStrategyType is a constant for spot strategy type.
	// HOLDING - Накопичуємо цільовий токен
	// SCALPING - Купуємо/продаемо цільовий токен за базовий
	// ARBITRAGE - Арбітраж, поки не реалізовано
	// TRADING - Трейдинг, накопичуємо цільовий токен, потім продаємо лімітним ордером
	// GRID - Грід, розміщуємо лімітні ордери на купівлю/продажу по сітці,
	// як спрацює ордер, ставимо новий, поки не вийдемо з позиції
	// Відслідковуємо рівень можливих втрат, якщо втрати перевищують ліміт, зупиняемо збільшення позиції
	// Коли ціна ліквідаціі починає наближатися, зменшуємо позицію
	// HOLDING/SCALPING/ARBITRAGE/TRADING/GRID
	HoldingStrategyType   StrategyType = "HOLDING"
	ScalpingStrategyType  StrategyType = "SCALPING"
	ArbitrageStrategyType StrategyType = "ARBITRAGE"
	TradingStrategyType   StrategyType = "TRADING"
	GridStrategyType      StrategyType = "GRID"
	GridStrategyTypeV2    StrategyType = "GRID_V2"
	GridStrategyTypeV3    StrategyType = "GRID_V3"
	GridStrategyTypeV4    StrategyType = "GRID_V4"
	GridStrategyTypeV5    StrategyType = "GRID_V5"
	// INPUT_INTO_POSITION - Режим входу - накопичуємо цільовий токен
	// WORK_IN_POSITION - Режим спекуляції - купуємо/продаемо цільовий токен за базовий
	// OUTPUT_OF_POSITION - Режим виходу - продаемо цільовий токен
	// SpotStageType is a constant for spot stage type.
	// INPUT_INTO_POSITION/WORK_IN_POSITION/OUTPUT_OF_POSITION/CLOSED
	InputIntoPositionStage StageType = "INPUT_INTO_POSITION"
	WorkInPositionStage    StageType = "WORK_IN_POSITION"
	OutputOfPositionStage  StageType = "OUTPUT_OF_POSITION"
	PositionClosedStage    StageType = "CLOSED"

	// Для USDT_FUTURE/COIN_FUTURE
	CrossMarginType    MarginType = "CROSS"
	IsolatedMarginType MarginType = "ISOLATED"

	// Арифметична прогресія
	ArithmeticProgression ProgressionType = "ARITHMETIC"
	// Геометрична прогресія
	GeometricProgression ProgressionType = "GEOMETRIC"
	// Експоненціальна прогресія
	ExponentialProgression ProgressionType = "EXPONENTIAL"
	// Логарифмічна прогресія
	LogarithmicProgression ProgressionType = "LOGARITHMIC"
	// Квадратична прогресія
	QuadraticProgression ProgressionType = "QUADRATIC"
	// Кубічна прогресія
	CubicProgression ProgressionType = "CUBIC"
	// Квадратно-коренева прогресія
	SquareRootProgression ProgressionType = "SQUARE_ROOT"
	// Кубічно-коренева прогресія
	CubicRootProgression ProgressionType = "CUBIC_ROOT"
	// Гармонічна прогресія
	HarmonicProgression ProgressionType = "HARMONIC"
)

type (
	AccountType     string
	MarginType      string
	ProgressionType string
	StageType       string
	StrategyType    string
	Pairs           struct {
		AccountType  AccountType  `json:"account_type"`  // Тип акаунта
		StrategyType StrategyType `json:"strategy_type"` // Тип стратегії
		StageType    StageType    `json:"stage_type"`    // Cтадія стратегії

		Pair string `json:"symbol"` // Пара

		// Для USDT_FUTURE/COIN_FUTURE
		MarginType MarginType `json:"margin_type"` // Тип маржі
		Leverage   int        `json:"leverage"`    // Маржинальне плече

		LimitOnPosition    float64 `json:"limit_on_position"`    // Ліміт на позицію, відсоток від балансу базової валюти
		LimitOnTransaction float64 `json:"limit_on_transaction"` // Ліміт на транзакцію, відсоток від ліміту на позицію

		UpBound  float64 `json:"up_bound"`  // Верхня межа ціни, най буде відсоток від ціни безубитку позиції
		LowBound float64 `json:"low_bound"` // Нижня межа ціни, най буде відсоток від ціни безубитку позиції
		MinSteps int     `json:"min_steps"` // Мінімальна кількість кроків

		DeltaPrice    float64         `json:"delta_price"`    // Дельта для купівлі/продажу
		DeltaQuantity float64         `json:"delta_quantity"` // Кількість для купівлі/продажу
		Progression   ProgressionType `json:"progression"`    // Тип прогресії

		Value float64 `json:"value"` // Вартість позиції

		CallbackRate float64 `json:"callback_rate"` // callbackRate для TRAILING_STOP_MARKET

		PercentToTarget float64 `json:"percent_to_target"` // Відсоток до цільової позиції
		PercentToLimit  float64 `json:"percent_to_limit"`  // Відсоток до ліміту на позицію
		DepthsN         int     `json:"depths_n"`          // Глибина стакана
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

// Get AccountType implements Pairs.
func (pr *Pairs) GetAccountType() AccountType {
	return pr.AccountType
}

// GetStrategy implements Pairs.
func (pr *Pairs) GetStrategy() StrategyType {
	return pr.StrategyType
}

// SetStrategy implements Pairs.
func (pr *Pairs) SetStrategy(strategy StrategyType) {
	pr.StrategyType = strategy
}

// GetStage implements Pairs.
func (pr *Pairs) GetStage() StageType {
	return pr.StageType
}

// SetStage implements Pairs.
func (pr *Pairs) SetStage(stage StageType) {
	pr.StageType = stage
}

// GetSymbol implements Pairs.
func (pr *Pairs) GetPair() string {
	return pr.Pair
}

// GetMarginType implements Pairs.
func (pr *Pairs) GetMarginType() MarginType {
	return pr.MarginType
}

// SetMarginType implements pairs.Pairs.
func (pr *Pairs) SetMarginType(marginType MarginType) {
	pr.MarginType = MarginType(strings.ToUpper(string(marginType)))
}

// SetMarginType implements Pairs.
func (pr *Pairs) GetLeverage() int {
	return pr.Leverage
}

// SetLeverage implements Pairs.
func (pr *Pairs) SetLeverage(leverage int) {
	pr.Leverage = leverage
}

func (pr *Pairs) GetLimitOnPosition() float64 {
	return pr.LimitOnPosition
}

func (pr *Pairs) GetLimitOnTransaction() float64 {
	return pr.LimitOnTransaction / 100
}

func (pr *Pairs) GetUpBound() float64 {
	return pr.UpBound / 100
}

func (pr *Pairs) GetLowBound() float64 {
	return pr.LowBound / 100
}

func (pr *Pairs) GetMinSteps() int {
	return pr.MinSteps
}

func (pr *Pairs) GetDeltaPrice() float64 {
	return pr.DeltaPrice / 100
}

func (pr *Pairs) GetDeltaQuantity() float64 {
	return pr.DeltaQuantity / 100
}

func (pr *Pairs) GetProgression() ProgressionType {
	return pr.Progression
}

func (pr *Pairs) GetValue() float64 {
	return pr.Value
}

func (pr *Pairs) SetLimitOnPosition(val float64) {
	pr.LimitOnPosition = val
}

func (pr *Pairs) SetLimitOnTransaction(val float64) {
	pr.LimitOnTransaction = val * 100
}

func (pr *Pairs) SetUpBoundPercent(val float64) {
	pr.UpBound = val * 100
}

func (pr *Pairs) SetLowBoundPercent(val float64) {
	pr.LowBound = val * 100
}

func (pr *Pairs) SetMinSteps(val int) {
	pr.MinSteps = val
}

func (pr *Pairs) SetDeltaPrice(val float64) {
	pr.DeltaPrice = val * 100
}

func (pr *Pairs) SetDeltaQuantity(quantity float64) {
	pr.DeltaQuantity = quantity * 100
}

func (pr *Pairs) SetProgression(val ProgressionType) {
	pr.Progression = val
}

func (pr *Pairs) SetValue(value float64) {
	pr.Value = value
}

func (pr *Pairs) GetCallbackRate() float64 {
	return pr.CallbackRate
}

func (pr *Pairs) SetCallbackRate(rate float64) {
	pr.CallbackRate = rate
}

func (pr *Pairs) GetPercentToTarget() float64 {
	if pr.PercentToTarget == 0 {
		return 10
	} else {
		return pr.PercentToTarget
	}
}

func (pr *Pairs) SetPercentToTarget(percent float64) {
	pr.PercentToTarget = percent
}

func (pr *Pairs) GetPercentToLimit() float64 {
	if pr.PercentToLimit == 0 {
		return 75
	} else {
		return pr.PercentToLimit
	}
}

func (pr *Pairs) SetPercentToLimit(percent float64) {
	pr.PercentToLimit = percent
}

func (pr *Pairs) GetDepthsN() int {
	if pr.DepthsN == 0 {
		return 50
	} else {
		return pr.DepthsN
	}
}

func (pr *Pairs) SetDepthsN(n int) {
	pr.DepthsN = n
}

func New(
	connection *connection_types.Connection,
	accountType AccountType,
	strategyType StrategyType,
	stageType StageType,
	pair string,
	targetSymbol string,
	baseSymbol string,
) *Pairs {
	return &Pairs{
		AccountType:        accountType,
		StrategyType:       strategyType,
		StageType:          stageType,
		Pair:               pair,
		LimitOnPosition:    100.0, // 100$
		LimitOnTransaction: 1.0,   // 1%
		DeltaPrice:         1.0,   // 1%
		DeltaQuantity:      10.0,  // 10%
		Progression:        "GEOMETRIC",
		Value:              0.0,
		CallbackRate:       0.1,  // 0.1%
		PercentToTarget:    10.0, // 10%
		PercentToLimit:     75.0, // 75%
		DepthsN:            50,   // 50
	}
}
