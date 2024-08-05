package types

import "github.com/google/btree"

type (
	UserDataEventReasonType string
	UserDataEventType       string
	OrderSide               string
	OrderType               string
	SideType                string
	TimeInForceType         string
	QuantityType            string
	OrderExecutionType      string
	OrderStatusType         string
	WorkingType             string
	PositionSideType        string
	DepthSide               string

	StreamFunction       func() (chan struct{}, chan struct{}, error)
	InitFunction         func() (err error)
	ErrorHandlerFunction func(err error)

	AccountType     string
	MarginType      string
	ProgressionType string
	StageType       string
	StrategyType    string

	OrderIdType int64
)

// Функції для btree.Btree
func (i OrderIdType) Less(than btree.Item) bool {
	return i < than.(OrderIdType)
}

func (i OrderIdType) Equal(than btree.Item) bool {
	return i == than.(OrderIdType)
}

const (
	DepthSideAsk DepthSide = "ASK"
	DepthSideBid DepthSide = "BID"
	SideTypeBuy  OrderSide = "BUY"
	SideTypeSell OrderSide = "SELL"
	SideTypeNone OrderSide = "NONE"
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
