package pairs

import (
	"fmt"
	"time"

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
	// HOLDING/SCALPING/ARBITRAGE/TRADING
	HoldingStrategyType   StrategyType = "HOLDING"
	ScalpingStrategyType  StrategyType = "SCALPING"
	ArbitrageStrategyType StrategyType = "ARBITRAGE"
	TradingStrategyType   StrategyType = "TRADING"
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
		Connection                 *connection_types.Connection `json:"connection"`
		AccountType                AccountType                  `json:"account_type"`                  // Тип акаунта
		StrategyType               StrategyType                 `json:"strategy_type"`                 // Тип стратегії
		StageType                  StageType                    `json:"stage_type"`                    // Cтадія стратегії
		Pair                       string                       `json:"symbol"`                        // Пара
		TargetSymbol               string                       `json:"target_symbol"`                 // Цільовий токен
		BaseSymbol                 string                       `json:"base_symbol"`                   // Базовий токен
		InitialBalance             float64                      `json:"initial_balance"`               // Початковий баланс
		SleepingTime               time.Duration                `json:"sleeping_time"`                 // Час сплячки, міллісекунди!!!
		TakingPositionSleepingTime time.Duration                `json:"taking_position_sleeping_time"` // Час сплячки при вході в позицію, хвилини!!!
		MiddlePrice                float64                      `json:"middle_price"`                  // Середня ціна купівлі по позиції

		// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
		// поки не наберемо цей ліміт, не можемо перейти до режиму спекуляціі
		LimitInputIntoPosition float64 `json:"limit_input_into_position"`

		// Ліміт на вихід з позиції, відсоток від балансу базової валюти,
		// як тільки наберемо цей ліміт, мусимо вийти з режиму спекуляціі
		// LimitOutputOfPosition > LimitInputIntoPosition
		LimitOutputOfPosition float64 `json:"limit_output_of_position"`

		LimitOnPosition    float64 `json:"limit_on_position"`    // Ліміт на позицію, відсоток від балансу базової валюти
		LimitOnTransaction float64 `json:"limit_on_transaction"` // Ліміт на транзакцію, відсоток від ліміту на позицію

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
func (cr *Pairs) Less(item btree.Item) bool {
	other := item.(*Pairs)
	if cr.AccountType != other.AccountType && cr.AccountType != "" && other.AccountType != "" {
		return cr.AccountType < other.AccountType
	}
	if cr.StrategyType != other.StrategyType && cr.StrategyType != "" && other.StrategyType != "" {
		return cr.StrategyType < other.StrategyType
	}
	if cr.StageType != other.StageType && cr.StageType != "" && other.StageType != "" {
		return cr.StageType < other.StageType
	}
	return cr.Pair < other.Pair
}

// Equals implements btree.Item.
func (cr *Pairs) Equals(item btree.Item) bool {
	other := item.(*Pairs)
	return cr.AccountType == other.AccountType &&
		cr.StrategyType == other.StrategyType &&
		cr.StageType == other.StageType &&
		cr.Pair == other.Pair
}

func (cr *Pairs) GetConnection() *connection_types.Connection {
	return cr.Connection
}

func (cr *Pairs) SetConnection(connection *connection_types.Connection) {
	cr.Connection = connection
}

// GetInitialBalance implements Configuration.
func (cr *Pairs) GetInitialBalance() float64 {
	return cr.InitialBalance
}

// SetInitialBalance implements Configuration.
func (cr *Pairs) SetInitialBalance(balance float64) {
	cr.InitialBalance = balance
}

// Get AccountType implements Configuration.
func (cr *Pairs) GetAccountType() AccountType {
	return cr.AccountType
}

// GetStrategy implements Configuration.
func (cr *Pairs) GetStrategy() StrategyType {
	return cr.StrategyType
}

// GetStage implements Configuration.
func (cr *Pairs) GetStage() StageType {
	return cr.StageType
}

// SetStage implements Configuration.
func (cr *Pairs) SetStage(stage StageType) {
	cr.StageType = stage
}

// GetSymbol implements Configuration.
func (cr *Pairs) GetPair() string {
	return cr.Pair
}

// GetTargetSymbol implements config.Configuration.
func (cr *Pairs) GetTargetSymbol() string {
	return cr.TargetSymbol
}

// GetBaseSymbol implements config.Configuration.
func (cr *Pairs) GetBaseSymbol() string {
	return cr.BaseSymbol
}

// GetSleepingTime implements Configuration.
func (cr *Pairs) GetSleepingTime() time.Duration {
	return cr.SleepingTime * time.Millisecond
}

// GetTakingPositionSleepingTime implements Configuration.
func (cr *Pairs) GetTakingPositionSleepingTime() time.Duration {
	return cr.TakingPositionSleepingTime * time.Minute
}

func (cr *Pairs) GetLimitInputIntoPosition() float64 {
	return cr.LimitInputIntoPosition
}

func (cr *Pairs) GetLimitOutputOfPosition() float64 {
	return cr.LimitOutputOfPosition
}

func (cr *Pairs) GetLimitOnPosition() float64 {
	return cr.LimitOnPosition
}

func (cr *Pairs) GetLimitOnTransaction() float64 {
	return cr.LimitOnTransaction
}

func (cr *Pairs) GetBuyDelta() float64 {
	return cr.BuyDelta
}

func (cr *Pairs) GetSellDelta() float64 {
	return cr.SellDelta
}

func (cr *Pairs) GetBuyQuantity() float64 {
	return cr.BuyQuantity
}

func (cr *Pairs) GetSellQuantity() float64 {
	return cr.SellQuantity
}

func (cr *Pairs) GetBuyValue() float64 {
	return cr.BuyValue
}

func (cr *Pairs) GetSellValue() float64 {
	return cr.SellValue
}

func (cr *Pairs) SetBuyQuantity(quantity float64) {
	cr.BuyQuantity = quantity
}

func (cr *Pairs) SetSellQuantity(quantity float64) {
	cr.SellQuantity = quantity
}

func (cr *Pairs) SetBuyValue(value float64) {
	cr.BuyValue = value
}

func (cr *Pairs) SetSellValue(value float64) {
	cr.SellValue = value
}

func (cr *Pairs) GetBuyCommission() float64 {
	return cr.BuyCommission
}

func (cr *Pairs) SetBuyCommission(commission float64) {
	cr.BuyCommission = commission
}

func (cr *Pairs) GetSellCommission() float64 {
	return cr.SellCommission
}

func (cr *Pairs) SetSellCommission(commission float64) {
	cr.SellCommission = commission
}

func (cr *Pairs) AddCommission(commission *binance.Fill) {
	cr.Commission[commission.CommissionAsset] += float64(utils.ConvStrToFloat64(commission.Commission))
}

func (cr *Pairs) GetCommission() Commission {
	return cr.Commission
}

func (cr *Pairs) SetCommission(commission Commission) {
	cr.Commission = commission
}

func (cr *Pairs) CalcMiddlePrice() error {
	if cr.BuyQuantity == cr.SellQuantity {
		return fmt.Errorf("BuyQuantity: %f and SellQuantity %f, can't calculate middle price", cr.BuyQuantity, cr.SellQuantity)
	}

	cr.MiddlePrice = (cr.BuyValue - cr.SellValue) / (cr.BuyQuantity - cr.SellQuantity)
	return nil
}

func (cr *Pairs) GetMiddlePrice() float64 {
	return cr.MiddlePrice
}

func (cr *Pairs) SetMiddlePrice(price float64) {
	cr.MiddlePrice = price
}

func (cr *Pairs) GetProfit(currentPrice float64) float64 {
	return (currentPrice - cr.GetMiddlePrice()) * (cr.BuyQuantity - cr.SellQuantity)
}

func (cr *Pairs) CheckingPair() bool {
	return cr.MiddlePrice != 0 &&
		cr.LimitInputIntoPosition != 0 &&
		cr.LimitOutputOfPosition != 0 &&
		cr.LimitInputIntoPosition < cr.LimitOutputOfPosition
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
		AccountType:            accountType,
		StrategyType:           strategyType,
		StageType:              stageType,
		Pair:                   pair,
		TargetSymbol:           targetSymbol,
		BaseSymbol:             baseSymbol,
		SleepingTime:           3 * time.Minute,
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
