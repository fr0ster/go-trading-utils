package pairs

import (
	"fmt"

	"github.com/adshao/go-binance/v2"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"

	"github.com/fr0ster/go-trading-utils/utils"

	"github.com/google/btree"
)

const (
	// SpotAccountType is a constant for spot account type.
	// SPOT/MARGIN/ISOLATED_MARGIN/USDT_FUTURE/COIN_FUTURE
	SpotAccountType    AccountType = "SPOT"
	MarginAccountType  AccountType = "MARGIN"
	IsolatedMarginType AccountType = "ISOLATED_MARGIN"
	USDTFutureType     AccountType = "USDT_FUTURE"
	CoinFutureType     AccountType = "COIN_FUTURE"
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
	// INPUT_INTO_POSITION - Режим входу - накопичуємо цільовий токен
	// WORK_IN_POSITION - Режим спекуляції - купуємо/продаемо цільовий токен за базовий
	// OUTPUT_OF_POSITION - Режим виходу - продаемо цільовий токен
	// SpotStageType is a constant for spot stage type.
	// INPUT_INTO_POSITION/WORK_IN_POSITION/OUTPUT_OF_POSITION/CLOSED
	InputIntoPositionStage StageType = "INPUT_INTO_POSITION"
	WorkInPositionStage    StageType = "WORK_IN_POSITION"
	OutputOfPositionStage  StageType = "OUTPUT_OF_POSITION"
	PositionClosedStage    StageType = "CLOSED"
)

type (
	AccountType  string
	StrategyType string
	StageType    string
	Commission   map[string]float64
	Pairs        struct {
		Connection             *connection_types.Connection `json:"connection"`
		AccountType            AccountType                  `json:"account_type"`             // Тип акаунта
		StrategyType           StrategyType                 `json:"strategy_type"`            // Тип стратегії
		StageType              StageType                    `json:"stage_type"`               // Cтадія стратегії
		Pair                   string                       `json:"symbol"`                   // Пара
		TargetSymbol           string                       `json:"target_symbol"`            // Цільовий токен
		BaseSymbol             string                       `json:"base_symbol"`              // Базовий токен
		InitialBalance         float64                      `json:"initial_balance"`          // Початковий баланс
		CurrentBalance         float64                      `json:"current_balance"`          // Поточний баланс
		InitialPositionBalance float64                      `json:"initial_position_balance"` // Початковий баланс позиції
		CurrentPositionBalance float64                      `json:"current_position_balance"` // Поточний баланс позиції
		MiddlePrice            float64                      `json:"middle_price"`             // Середня ціна купівлі по позиції

		// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
		// поки не наберемо цей ліміт, не можемо перейти до режиму спекуляціі
		LimitInputIntoPosition float64 `json:"limit_input_into_position"`

		// Ліміт на вихід з позиції, відсоток від балансу базової валюти,
		// як тільки наберемо цей ліміт, мусимо вийти з режиму спекуляціі
		// LimitOutputOfPosition > LimitInputIntoPosition
		LimitOutputOfPosition float64 `json:"limit_output_of_position"`

		LimitOnPosition    float64 `json:"limit_on_position"`    // Ліміт на позицію, відсоток від балансу базової валюти
		LimitOnTransaction float64 `json:"limit_on_transaction"` // Ліміт на транзакцію, відсоток від ліміту на позицію

		UpBound  float64 `json:"up_bound"`  // Верхня межа
		LowBound float64 `json:"low_bound"` // Нижня межа

		BuyDelta       float64            `json:"buy_delta"`       // Дельта для купівлі
		BuyQuantity    float64            `json:"buy_quantity"`    // Кількість для купівлі, суммарно по позиції
		BuyValue       float64            `json:"buy_value"`       // Вартість для купівлі, суммарно по позиції
		BuyCommission  float64            `json:"buy_commission"`  // Комісія за купівлю
		SellDelta      float64            `json:"sell_delta"`      // Дельта для продажу, суммарно по позиції
		SellQuantity   float64            `json:"sell_quantity"`   // Кількість для продажу, суммарно по позиції
		SellValue      float64            `json:"sell_value"`      // Вартість для продажу, суммарно по позиції
		SellCommission float64            `json:"sell_commission"` // Комісія за продаж
		Commission     map[string]float64 `json:"commission"`      // Комісія
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

func (pr *Pairs) GetConnection() *connection_types.Connection {
	return pr.Connection
}

func (pr *Pairs) SetConnection(connection *connection_types.Connection) {
	pr.Connection = connection
}

// GetInitialBalance implements Pairs.
func (pr *Pairs) GetInitialBalance() float64 {
	return pr.InitialBalance
}

// SetInitialBalance implements Pairs.
func (pr *Pairs) SetInitialBalance(balance float64) {
	pr.InitialBalance = balance
}

// GetCurrentBalance implements Pairs.
func (pr *Pairs) GetCurrentBalance() float64 {
	return pr.CurrentBalance
}

// SetCurrentBalance implements Pairs.
func (pr *Pairs) SetCurrentBalance(balance float64) {
	pr.CurrentBalance = balance
}

// GetInitialPositionBalance implements Pairs.
func (pr *Pairs) GetInitialPositionBalance() float64 {
	return pr.InitialPositionBalance
}

// SetInitialPositionBalance implements Pairs.
func (pr *Pairs) SetInitialPositionBalance(balance float64) {
	pr.InitialPositionBalance = balance
}

// GetCurrentPositionBalance implements Pairs.
func (pr *Pairs) GetCurrentPositionBalance() float64 {
	return pr.CurrentPositionBalance
}

// SetCurrentPositionBalance implements Pairs.
func (pr *Pairs) SetCurrentPositionBalance(balance float64) {
	pr.CurrentPositionBalance = balance
}

// Get AccountType implements Pairs.
func (pr *Pairs) GetAccountType() AccountType {
	return pr.AccountType
}

// GetStrategy implements Pairs.
func (pr *Pairs) GetStrategy() StrategyType {
	return pr.StrategyType
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

// GetTargetSymbol implements config.Configuration.
func (pr *Pairs) GetTargetSymbol() string {
	return pr.TargetSymbol
}

// GetBaseSymbol implements config.Configuration.
func (pr *Pairs) GetBaseSymbol() string {
	return pr.BaseSymbol
}

func (pr *Pairs) GetLimitInputIntoPosition() float64 {
	return pr.LimitInputIntoPosition
}

func (pr *Pairs) GetLimitOutputOfPosition() float64 {
	return pr.LimitOutputOfPosition
}

func (pr *Pairs) GetLimitOnPosition() float64 {
	return pr.LimitOnPosition
}

func (pr *Pairs) GetLimitOnTransaction() float64 {
	return pr.LimitOnTransaction
}

func (pr *Pairs) GetUpBound() float64 {
	return pr.UpBound
}

func (pr *Pairs) GetLowBound() float64 {
	return pr.LowBound
}

func (pr *Pairs) GetBuyDelta() float64 {
	return pr.BuyDelta
}

func (pr *Pairs) GetSellDelta() float64 {
	return pr.SellDelta
}

func (pr *Pairs) GetBuyQuantity() float64 {
	return pr.BuyQuantity
}

func (pr *Pairs) GetSellQuantity() float64 {
	return pr.SellQuantity
}

func (pr *Pairs) GetBuyValue() float64 {
	return pr.BuyValue
}

func (pr *Pairs) GetSellValue() float64 {
	return pr.SellValue
}

func (pr *Pairs) SetBuyQuantity(quantity float64) {
	pr.BuyQuantity = quantity
}

func (pr *Pairs) SetSellQuantity(quantity float64) {
	pr.SellQuantity = quantity
}

func (pr *Pairs) SetBuyValue(value float64) {
	pr.BuyValue = value
}

func (pr *Pairs) SetSellValue(value float64) {
	pr.SellValue = value
}

func (pr *Pairs) GetBuyCommission() float64 {
	return pr.BuyCommission
}

func (pr *Pairs) SetBuyCommission(commission float64) {
	pr.BuyCommission = commission
}

func (pr *Pairs) GetSellCommission() float64 {
	return pr.SellCommission
}

func (pr *Pairs) SetSellCommission(commission float64) {
	pr.SellCommission = commission
}

func (pr *Pairs) SetBuyData(quantity, value, commission float64) {
	pr.BuyQuantity = quantity
	pr.BuyValue = value
	pr.BuyCommission = commission
}

func (pr *Pairs) SetSellData(quantity, value, commission float64) {
	pr.SellQuantity = quantity
	pr.SellValue = value
	pr.SellCommission = commission
}

func (pr *Pairs) AddCommission(commission *binance.Fill) {
	pr.Commission[commission.CommissionAsset] += float64(utils.ConvStrToFloat64(commission.Commission))
}

func (pr *Pairs) GetCommission() Commission {
	return pr.Commission
}

func (pr *Pairs) SetCommission(commission Commission) {
	pr.Commission = commission
}

func (pr *Pairs) CalcMiddlePrice() error {
	if pr.BuyQuantity == pr.SellQuantity {
		return fmt.Errorf("BuyQuantity: %f and SellQuantity %f, can't calculate middle price", pr.BuyQuantity, pr.SellQuantity)
	}

	pr.MiddlePrice = (pr.BuyValue - pr.SellValue) / (pr.BuyQuantity - pr.SellQuantity)
	return nil
}

func (pr *Pairs) GetMiddlePrice() float64 {
	return pr.MiddlePrice
}

func (pr *Pairs) SetMiddlePrice(price float64) {
	pr.MiddlePrice = price
}

func (pr *Pairs) GetProfit(currentPrice float64) float64 {
	return (currentPrice - pr.GetMiddlePrice()) * (pr.BuyQuantity - pr.SellQuantity)
}

func (pr *Pairs) CheckingPair() bool {
	return pr.MiddlePrice != 0 &&
		pr.LimitInputIntoPosition != 0 &&
		pr.LimitOutputOfPosition != 0 &&
		pr.LimitInputIntoPosition < pr.LimitOutputOfPosition
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
		Connection:             connection,
		InitialBalance:         0.0,
		InitialPositionBalance: 0.0,
		AccountType:            accountType,
		StrategyType:           strategyType,
		StageType:              stageType,
		Pair:                   pair,
		TargetSymbol:           targetSymbol,
		BaseSymbol:             baseSymbol,
		LimitInputIntoPosition: 0.1,
		LimitOutputOfPosition:  0.5,
		LimitOnPosition:        1.0,
		LimitOnTransaction:     0.01,
		BuyDelta:               0.01,
		BuyQuantity:            0.0,
		BuyValue:               0.0,
		SellDelta:              0.05,
		SellQuantity:           0.0,
		SellValue:              0.0,
		Commission:             Commission{},
	}
}
