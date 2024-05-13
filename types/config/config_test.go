package config_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
	"github.com/sirupsen/logrus"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

const (
	SpotAPIKey        = "your_api_key"    // Ключ API для спот-біржі
	SpotAPISecret     = "your_api_secret" // Секретний ключ API для спот-біржі
	SpotUseTestNet    = false             // Використовувати тестову мережу для спот-біржі
	FuturesAPIKey     = "your_api_key"    // Ключ API для ф'ючерсів
	FuturesAPISecret  = "your_api_secret" // Секретний ключ API для ф'ючерсів
	FuturesUseTestNet = false             // Використовувати тестову мережу для ф'ючерсів
	InfoLevel         = logrus.InfoLevel  // Рівень логування
	DebugLevel        = logrus.DebugLevel // Рівень логування

	InitialBalance = 1000.0 // Початковий баланс
	CurrentBalance = 2000.0 // Поточний баланс

	SpotCommissionMaker = 0.001 // Комісія за мейкером
	SpotCommissionTaker = 0.001 // Комісія за тейкером

	AccountType_1                = pairs_types.SpotAccountType        // Тип акаунта
	StrategyType_1               = pairs_types.HoldingStrategyType    // Тип стратегії
	StageType_1                  = pairs_types.InputIntoPositionStage // Стадія стратегії
	Pair_1                       = "BTCUSDT"                          // Пара
	TargetSymbol_1               = "BTC"                              // Котирувальна валюта
	BaseSymbol_1                 = "USDT"                             // Базова валюта
	BaseBalance_1                = 2000.0                             // Баланс базової валюти
	TargetBalance_1              = 1000.0                             // Баланс цільової валюти
	MiddlePrice_1                = 40000.0                            // Середня ціна купівлі по позиції
	SleepingTime_1               = 5                                  // Час сплячки, міллісекунди
	TakingPositionSleepingTime_1 = 60                                 // Час сплячки при вході в позицію, хвилини

	// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
	// повинно бути меньше ніж LimitOutputOfPosition_1,
	// але це для перевірки CheckingPair
	LimitInputIntoPosition_1 = 0.01
	// Ліміт на вихід з позиції, відсоток від балансу базової валюти,
	// повинно бути більше ніж LimitInputIntoPosition_1,
	// але це для перевірки CheckingPair
	LimitOutputOfPosition_1 = 0.05

	LimitOnPosition_1        = 0.50                               // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_1     = 0.10                               // Ліміт на транзакцію, відсоток від ліміту на позицію
	InitialPositionBalance_1 = InitialBalance * LimitOnPosition_1 // Початковий баланс позиції
	CurrentPositionBalance_1 = CurrentBalance * LimitOnPosition_1 // Поточний баланс позиції

	BuyDelta_1       = 0.01  // Дельта для купівлі
	SellDelta_1      = 0.01  // Дельта для продажу
	BuyQuantity_1    = 1.0   // Кількість для купівлі, суммарно по позиції
	BuyCommission_1  = 0.001 // Комісія за купівлю
	SellQuantity_1   = 1.0   // Кількість для продажу, суммарно по позиції
	BuyValue_1       = 100.0 // Вартість для купівлі, суммарно по позиції
	SellValue_1      = 100.0 // Вартість для продажу, суммарно по позиції
	SellCommission_1 = 0.001 // Комісія за продаж

	FuturesCommissionMaker = 0.002 // Комісія за мейкером
	FuturesCommissionTaker = 0.005 // Комісія за тейкером

	AccountType_2   = pairs_types.USDTFutureType      // Тип акаунта
	StrategyType_2  = pairs_types.TradingStrategyType // Тип стратегії
	StageType_2     = pairs_types.WorkInPositionStage // Тип стадії
	Pair_2          = "ETHUSDT"                       // Пара
	TargetSymbol_2  = "ETH"                           // Котирувальна валюта
	BaseSymbol_2    = "USDT"                          // Базова валюта
	BaseBalance_2   = 2000.0                          // Баланс базової валюти
	TargetBalance_2 = 10.0                            // Баланс цільової валюти
	MiddlePrice_2   = 3000.0                          // Середня ціна купівлі по позиції

	// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
	// повинно бути меньше ніж LimitOutputOfPosition_2,
	// але це для перевірки CheckingPair
	LimitInputIntoPosition_2 = 0.15
	// Ліміт на вихід з позиції, відсоток від балансу базової валюти,
	// повинно бути більше ніж LimitInputIntoPosition_2,
	// але це для перевірки CheckingPair
	LimitOutputOfPosition_2 = 0.10 // Ліміт на вихід з позиції, відсоток від балансу базової валюти

	LimitOnPosition_2        = 0.50                               // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_2     = 0.01                               // Ліміт на транзакцію, відсоток від ліміту на позицію
	InitialPositionBalance_2 = InitialBalance * LimitOnPosition_2 // Початковий баланс позиції
	CurrentPositionBalance_2 = CurrentBalance * LimitOnPosition_2 // Поточний баланс позиції

	BuyDelta_2       = 0.01   // Дельта для купівлі
	SellDelta_2      = 0.01   // Дельта для продажу
	BuyQuantity_2    = 1.0    // Кількість для купівлі, суммарно по позиції
	BuyCommission_2  = 0.0002 // Комісія за купівлю
	SellQuantity_2   = 1.0    // Кількість для продажу, суммарно по позиції
	BuyValue_2       = 100.0  // Вартість для купівлі, суммарно по позиції
	SellValue_2      = 100.0  // Вартість для продажу, суммарно по позиції
	SellCommission_2 = 0.0002 // Комісія за продаж
)

var (
	CommissionAsset_1     = "BNB"
	CommissionAsset_2     = "USDT"
	Commission_1          = 0.00001
	Commission_2          = 0.02
	Commission            = pairs_types.Commission{CommissionAsset_1: Commission_1, CommissionAsset_2: Commission_2}
	DefaultSpotConnection = &connection_types.Connection{
		APIKey:          SpotAPIKey,
		APISecret:       SpotAPISecret,
		UseTestNet:      SpotUseTestNet,
		CommissionMaker: SpotCommissionMaker,
		CommissionTaker: SpotCommissionTaker,
	}
	DefaultFuturesConnection = &connection_types.Connection{
		APIKey:          FuturesAPIKey,
		APISecret:       FuturesAPISecret,
		UseTestNet:      FuturesUseTestNet,
		CommissionMaker: FuturesCommissionMaker,
		CommissionTaker: FuturesCommissionTaker,
	}
	config = &config_types.Configs{
		SpotConnection: &connection_types.Connection{
			APIKey:          SpotAPIKey,
			APISecret:       SpotAPISecret,
			UseTestNet:      SpotUseTestNet,
			CommissionMaker: SpotCommissionMaker,
			CommissionTaker: SpotCommissionTaker,
		},
		FuturesConnection: &connection_types.Connection{
			APIKey:          FuturesAPIKey,
			APISecret:       FuturesAPISecret,
			UseTestNet:      FuturesUseTestNet,
			CommissionMaker: FuturesCommissionMaker,
			CommissionTaker: FuturesCommissionTaker,
		},
		LogLevel: InfoLevel,
		Pairs:    btree.New(2),
	}
	pair_1 = &pairs_types.Pairs{
		Connection:             &connection_types.Connection{},
		InitialBalance:         InitialBalance,
		CurrentBalance:         CurrentBalance,
		InitialPositionBalance: InitialPositionBalance_1,
		CurrentPositionBalance: CurrentPositionBalance_1,
		AccountType:            AccountType_1,
		StrategyType:           StrategyType_1,
		StageType:              StageType_1,
		Pair:                   Pair_1,
		TargetSymbol:           TargetSymbol_1,
		BaseSymbol:             BaseSymbol_1,
		MiddlePrice:            MiddlePrice_1,
		LimitInputIntoPosition: LimitInputIntoPosition_1,
		LimitOutputOfPosition:  LimitOutputOfPosition_1,
		LimitOnPosition:        LimitOnPosition_1,
		LimitOnTransaction:     LimitOnTransaction_1,
		BuyDelta:               BuyDelta_1,
		BuyQuantity:            BuyQuantity_1,
		BuyValue:               BuyValue_1,
		BuyCommission:          BuyCommission_1,
		SellDelta:              SellDelta_1,
		SellQuantity:           SellQuantity_1,
		SellValue:              SellValue_1,
		SellCommission:         SellCommission_1,
		Commission:             Commission,
	}
	pair_2 = &pairs_types.Pairs{
		Connection:             &connection_types.Connection{},
		InitialBalance:         InitialBalance,
		CurrentBalance:         CurrentBalance,
		InitialPositionBalance: InitialPositionBalance_2,
		CurrentPositionBalance: CurrentPositionBalance_2,
		AccountType:            AccountType_2,
		StrategyType:           StrategyType_2,
		StageType:              StageType_2,
		Pair:                   Pair_2,
		TargetSymbol:           TargetSymbol_2,
		BaseSymbol:             BaseSymbol_2,
		MiddlePrice:            MiddlePrice_2,
		LimitInputIntoPosition: LimitInputIntoPosition_2,
		LimitOutputOfPosition:  LimitOutputOfPosition_2,
		LimitOnPosition:        LimitOnPosition_2,
		LimitOnTransaction:     LimitOnTransaction_2,
		BuyDelta:               BuyDelta_2,
		BuyQuantity:            BuyQuantity_2,
		BuyValue:               BuyValue_2,
		BuyCommission:          BuyCommission_2,
		SellDelta:              SellDelta_2,
		SellQuantity:           SellQuantity_2,
		SellValue:              SellValue_2,
		SellCommission:         SellCommission_2,
		Commission:             Commission,
	}
)

func getTestData() []byte {
	return []byte(
		`{
			"spot_connection": {
				"api_key": "` + SpotAPIKey + `",
				"api_secret": "` + SpotAPISecret + `",
				"use_test_net": ` + strconv.FormatBool(SpotUseTestNet) + `,
				"commission_maker": ` + json.Number(strconv.FormatFloat(SpotCommissionMaker, 'f', -1, 64)).String() + `,
				"commission_taker": ` + json.Number(strconv.FormatFloat(SpotCommissionTaker, 'f', -1, 64)).String() + `
			},
			"futures_connection": {
				"api_key": "` + FuturesAPIKey + `",
				"api_secret": "` + FuturesAPISecret + `",
				"use_test_net": ` + strconv.FormatBool(FuturesUseTestNet) + `,
				"commission_maker": ` + json.Number(strconv.FormatFloat(FuturesCommissionMaker, 'f', -1, 64)).String() + `,
				"commission_taker": ` + json.Number(strconv.FormatFloat(FuturesCommissionTaker, 'f', -1, 64)).String() + `
			},
			"log_level": "` + InfoLevel.String() + `",
			"pairs": [
				{
					"connection": {
						"api_key": "` + SpotAPIKey + `",
						"api_secret": "` + SpotAPISecret + `",
						"use_test_net": ` + strconv.FormatBool(SpotUseTestNet) + `,
						"commission_maker": ` + json.Number(strconv.FormatFloat(SpotCommissionMaker, 'f', -1, 64)).String() + `,
						"commission_taker": ` + json.Number(strconv.FormatFloat(SpotCommissionTaker, 'f', -1, 64)).String() + `
					},
					"initial_balance": ` + json.Number(strconv.FormatFloat(InitialBalance, 'f', -1, 64)).String() + `,
					"current_balance": ` + json.Number(strconv.FormatFloat(CurrentBalance, 'f', -1, 64)).String() + `,
					"initial_position_balance": ` + json.Number(strconv.FormatFloat(InitialPositionBalance_1, 'f', -1, 64)).String() + `,
					"current_position_balance": ` + json.Number(strconv.FormatFloat(CurrentPositionBalance_1, 'f', -1, 64)).String() + `,
					"account_type": "` + string(AccountType_1) + `",
					"strategy_type": "` + string(StrategyType_1) + `",
					"stage_type": "` + string(StageType_1) + `",
					"symbol": "` + Pair_1 + `",
					"target_symbol": "` + TargetSymbol_1 + `",
					"base_symbol": "` + BaseSymbol_1 + `",
					"sleeping_time": ` + strconv.Itoa(SleepingTime_1) + `,
					"taking_position_sleeping_time": ` + strconv.Itoa(TakingPositionSleepingTime_1) + `,
					"middle_price": ` + json.Number(strconv.FormatFloat(MiddlePrice_1, 'f', -1, 64)).String() + `,
					"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitInputIntoPosition_1, 'f', -1, 64)).String() + `,
					"limit_output_of_position": ` + json.Number(strconv.FormatFloat(LimitOutputOfPosition_1, 'f', -1, 64)).String() + `,
					"limit_on_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_1, 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_1, 'f', -1, 64)).String() + `,
					"buy_delta": ` + json.Number(strconv.FormatFloat(BuyDelta_1, 'f', -1, 64)).String() + `,
					"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_1, 'f', -1, 64)).String() + `,
					"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_1, 'f', -1, 64)).String() + `,
					"buy_commission": ` + json.Number(strconv.FormatFloat(BuyCommission_1, 'f', -1, 64)).String() + `,
					"sell_delta": ` + json.Number(strconv.FormatFloat(SellDelta_1, 'f', -1, 64)).String() + `,
					"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_1, 'f', -1, 64)).String() + `,
					"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_1, 'f', -1, 64)).String() + `,
					"sell_commission": ` + json.Number(strconv.FormatFloat(SellCommission_1, 'f', -1, 64)).String() + `,
					"commission": {
						"` + CommissionAsset_1 + `": ` + json.Number(strconv.FormatFloat(Commission_1, 'f', -1, 64)).String() + `,
						"` + CommissionAsset_2 + `": ` + json.Number(strconv.FormatFloat(Commission_2, 'f', -1, 64)).String() + `
					}
				},
				{
					"connection": {
						"api_key": "` + FuturesAPIKey + `",
						"api_secret": "` + FuturesAPISecret + `",
						"use_test_net": ` + strconv.FormatBool(FuturesUseTestNet) + `,
						"commission_maker": ` + json.Number(strconv.FormatFloat(FuturesCommissionMaker, 'f', -1, 64)).String() + `,
						"commission_taker": ` + json.Number(strconv.FormatFloat(FuturesCommissionTaker, 'f', -1, 64)).String() + `
					},
					"initial_balance": ` + json.Number(strconv.FormatFloat(InitialBalance, 'f', -1, 64)).String() + `,
					"current_balance": ` + json.Number(strconv.FormatFloat(CurrentBalance, 'f', -1, 64)).String() + `,
					"initial_position_balance": ` + json.Number(strconv.FormatFloat(InitialPositionBalance_2, 'f', -1, 64)).String() + `,
					"current_position_balance": ` + json.Number(strconv.FormatFloat(CurrentPositionBalance_2, 'f', -1, 64)).String() + `,
					"account_type": "` + string(AccountType_2) + `",
					"strategy_type": "` + string(StrategyType_2) + `",
					"stage_type": "` + string(StageType_2) + `",
					"symbol": "` + Pair_2 + `",
					"target_symbol": "` + TargetSymbol_2 + `",
					"base_symbol": "` + BaseSymbol_2 + `",
					"middle_price": ` + json.Number(strconv.FormatFloat(MiddlePrice_2, 'f', -1, 64)).String() + `,
					"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitInputIntoPosition_2, 'f', -1, 64)).String() + `,
					"limit_output_of_position": ` + json.Number(strconv.FormatFloat(LimitOutputOfPosition_2, 'f', -1, 64)).String() + `,
					"limit_in_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_2, 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_2, 'f', -1, 64)).String() + `,
					"buy_delta": ` + json.Number(strconv.FormatFloat(BuyDelta_2, 'f', -1, 64)).String() + `,
					"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_2, 'f', -1, 64)).String() + `,
					"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_2, 'f', -1, 64)).String() + `,
					"buy_commission": ` + json.Number(strconv.FormatFloat(BuyCommission_2, 'f', -1, 64)).String() + `,
					"sell_delta": ` + json.Number(strconv.FormatFloat(SellDelta_1, 'f', -1, 64)).String() + `,
					"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_2, 'f', -1, 64)).String() + `,
					"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_2, 'f', -1, 64)).String() + `,
					"sell_commission": ` + json.Number(strconv.FormatFloat(SellCommission_2, 'f', -1, 64)).String() + `,
					"commission": {
						"` + CommissionAsset_1 + `": ` + json.Number(strconv.FormatFloat(Commission_1, 'f', -1, 64)).String() + `,
						"` + CommissionAsset_2 + `": ` + json.Number(strconv.FormatFloat(Commission_2, 'f', -1, 64)).String() + `
					}
				}
			]
		}`)
}

func assertTest(t *testing.T, err error, config config_interfaces.Configuration, checkingDate *[]pairs_interfaces.Pairs) {
	assert.NoError(t, err)
	assert.Equal(t, SpotAPIKey, config.GetSpotConnection().GetAPIKey())
	assert.Equal(t, SpotAPISecret, config.GetSpotConnection().GetSecretKey())
	assert.Equal(t, SpotUseTestNet, config.GetSpotConnection().GetUseTestNet())
	assert.Equal(t, FuturesAPIKey, config.GetFuturesConnection().GetAPIKey())
	assert.Equal(t, FuturesAPISecret, config.GetFuturesConnection().GetSecretKey())
	assert.Equal(t, FuturesUseTestNet, config.GetFuturesConnection().GetUseTestNet())
	assert.Equal(t, SpotCommissionMaker, config.GetSpotConnection().GetCommissionMaker())
	assert.Equal(t, SpotCommissionTaker, config.GetSpotConnection().GetCommissionTaker())
	assert.Equal(t, InfoLevel, config.GetLogLevel())

	assert.Equal(t, (*checkingDate)[0].GetConnection().GetAPIKey(), config.GetPair(Pair_1).GetConnection().GetAPIKey())
	assert.Equal(t, (*checkingDate)[0].GetConnection().GetSecretKey(), config.GetPair(Pair_1).GetConnection().GetSecretKey())
	assert.Equal(t, (*checkingDate)[0].GetConnection().GetUseTestNet(), config.GetPair(Pair_1).GetConnection().GetUseTestNet())
	assert.Equal(t, (*checkingDate)[0].GetConnection().GetCommissionMaker(), config.GetPair(Pair_1).GetConnection().GetCommissionMaker())
	assert.Equal(t, (*checkingDate)[0].GetConnection().GetCommissionTaker(), config.GetPair(Pair_1).GetConnection().GetCommissionTaker())

	assert.Equal(t, (*checkingDate)[0].GetInitialBalance(), config.GetPair(Pair_1).GetInitialBalance())
	assert.Equal(t, (*checkingDate)[0].GetCurrentBalance(), config.GetPair(Pair_1).GetCurrentBalance())
	assert.Equal(t, (*checkingDate)[0].GetInitialPositionBalance(), config.GetPair(Pair_1).GetInitialPositionBalance())
	assert.Equal(t, (*checkingDate)[0].GetCurrentPositionBalance(), config.GetPair(Pair_1).GetCurrentPositionBalance())
	assert.Equal(t, (*checkingDate)[0].GetAccountType(), config.GetPair(Pair_1).GetAccountType())
	assert.Equal(t, (*checkingDate)[0].GetStrategy(), config.GetPair(Pair_1).GetStrategy())
	assert.Equal(t, (*checkingDate)[0].GetStage(), config.GetPair(Pair_1).GetStage())
	assert.Equal(t, (*checkingDate)[0].GetPair(), config.GetPair(Pair_1).GetPair())
	assert.Equal(t, (*checkingDate)[0].GetTargetSymbol(), config.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[0].GetBaseSymbol(), config.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[0].GetMiddlePrice(), config.GetPair(Pair_1).GetMiddlePrice())
	assert.Equal(t, (*checkingDate)[0].GetLimitInputIntoPosition(), config.GetPair(Pair_1).GetLimitInputIntoPosition())
	assert.Equal(t, (*checkingDate)[0].GetLimitOutputOfPosition(), config.GetPair(Pair_1).GetLimitOutputOfPosition())
	assert.Equal(t, (*checkingDate)[0].GetLimitOnPosition(), config.GetPair(Pair_1).GetLimitOnPosition())
	assert.Equal(t, (*checkingDate)[0].GetLimitOnTransaction(), config.GetPair(Pair_1).GetLimitOnTransaction())
	assert.Equal(t, (*checkingDate)[0].GetBuyDelta(), config.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[0].GetBuyQuantity(), config.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[0].GetBuyValue(), config.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, (*checkingDate)[0].GetBuyCommission(), config.GetPair(Pair_1).GetBuyCommission())
	assert.Equal(t, (*checkingDate)[0].GetSellDelta(), config.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, (*checkingDate)[0].GetSellQuantity(), config.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[0].GetSellValue(), config.GetPair(Pair_1).GetSellValue())
	assert.Equal(t, (*checkingDate)[0].GetSellCommission(), config.GetPair(Pair_1).GetSellCommission())
	assert.Equal(t, (*checkingDate)[0].GetCommission(), Commission)

	assert.Equal(t, (*checkingDate)[1].GetConnection().GetAPIKey(), config.GetPair(Pair_2).GetConnection().GetAPIKey())
	assert.Equal(t, (*checkingDate)[1].GetConnection().GetSecretKey(), config.GetPair(Pair_2).GetConnection().GetSecretKey())
	assert.Equal(t, (*checkingDate)[1].GetConnection().GetUseTestNet(), config.GetPair(Pair_2).GetConnection().GetUseTestNet())
	assert.Equal(t, (*checkingDate)[1].GetConnection().GetCommissionMaker(), config.GetPair(Pair_2).GetConnection().GetCommissionMaker())
	assert.Equal(t, (*checkingDate)[1].GetConnection().GetCommissionTaker(), config.GetPair(Pair_2).GetConnection().GetCommissionTaker())

	assert.Equal(t, (*checkingDate)[1].GetInitialBalance(), config.GetPair(Pair_2).GetInitialBalance())
	assert.Equal(t, (*checkingDate)[1].GetCurrentBalance(), config.GetPair(Pair_2).GetCurrentBalance())
	assert.Equal(t, (*checkingDate)[1].GetInitialPositionBalance(), config.GetPair(Pair_2).GetInitialPositionBalance())
	assert.Equal(t, (*checkingDate)[1].GetCurrentPositionBalance(), config.GetPair(Pair_2).GetCurrentPositionBalance())
	assert.Equal(t, (*checkingDate)[1].GetAccountType(), config.GetPair(Pair_2).GetAccountType())
	assert.Equal(t, (*checkingDate)[1].GetStrategy(), config.GetPair(Pair_2).GetStrategy())
	assert.Equal(t, (*checkingDate)[1].GetStage(), config.GetPair(Pair_2).GetStage())
	assert.Equal(t, (*checkingDate)[1].GetPair(), config.GetPair(Pair_2).GetPair())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), config.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), config.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetMiddlePrice(), config.GetPair(Pair_2).GetMiddlePrice())
	assert.Equal(t, (*checkingDate)[1].GetLimitInputIntoPosition(), config.GetPair(Pair_2).GetLimitInputIntoPosition())
	assert.Equal(t, (*checkingDate)[1].GetLimitOutputOfPosition(), config.GetPair(Pair_2).GetLimitOutputOfPosition())
	assert.Equal(t, (*checkingDate)[1].GetLimitOnPosition(), config.GetPair(Pair_2).GetLimitOnPosition())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), config.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), config.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetLimitOnTransaction(), config.GetPair(Pair_2).GetLimitOnTransaction())
	assert.Equal(t, (*checkingDate)[1].GetBuyDelta(), config.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[1].GetBuyQuantity(), config.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[1].GetBuyValue(), config.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, (*checkingDate)[1].GetBuyCommission(), config.GetPair(Pair_2).GetBuyCommission())
	assert.Equal(t, (*checkingDate)[1].GetSellDelta(), config.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, (*checkingDate)[1].GetSellQuantity(), config.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[1].GetSellValue(), config.GetPair(Pair_2).GetSellValue())
	assert.Equal(t, (*checkingDate)[1].GetSellCommission(), config.GetPair(Pair_2).GetSellCommission())
	assert.Equal(t, (*checkingDate)[1].GetCommission(), Commission)

}

func TestConfigFile_Load(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write test data to the temporary config file
	testData := getTestData()
	err = os.WriteFile(tmpFile.Name(), testData, 0644)
	assert.NoError(t, err)

	// Create a new ConfigFile instance
	configFile := config_types.NewConfigFile(tmpFile.Name(), 2)
	configFile.SetConfigurations(config)

	// Load the config from the file
	err = configFile.Load()
	assert.NoError(t, err)

	// Assert that the loaded config matches the test data
	checkingDate, err := configFile.GetConfigurations().GetPairs()
	assertTest(t, err, configFile.GetConfigurations(), checkingDate)
}

func TestConfigFile_Save(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a new ConfigFile instance
	config_file := config_types.NewConfigFile(tmpFile.Name(), 2)
	config_file.SetConfigurations(config)
	config_file.GetConfigurations().SetPair(pair_1)
	config_file.GetConfigurations().SetPair(pair_2)

	// Save the config to the file
	err = config_file.Save()
	assert.NoError(t, err)

	// Read the saved config file
	savedData, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	// Unmarshal the saved data into a ConfigFile struct
	savedConfig := &config_types.Configs{}
	err = json.Unmarshal(savedData, savedConfig)
	assert.NoError(t, err)

	// Assert that the saved config matches the original config
	checkingDate, err := config_file.GetConfigurations().GetPairs()
	assertTest(t, err, config_file.GetConfigurations(), checkingDate)
}

// Add more tests for other methods if needed

func TestPairSetter(t *testing.T) {
	pair := &pairs_types.Pairs{
		Pair:        Pair_1,
		BuyQuantity: BuyQuantity_1,
		BuyValue:    BuyValue_1,
		Commission:  Commission,
	}
	newCommission := pairs_types.Commission{"USDT": 0.1}
	pair.SetConnection(&connection_types.Connection{
		APIKey:          SpotAPIKey,
		APISecret:       SpotAPISecret,
		UseTestNet:      SpotUseTestNet,
		CommissionMaker: SpotCommissionMaker,
		CommissionTaker: SpotCommissionTaker,
	})
	assert.Equal(t, SpotAPIKey, pair.GetConnection().GetAPIKey())
	assert.Equal(t, SpotAPISecret, pair.GetConnection().GetSecretKey())
	assert.Equal(t, SpotUseTestNet, pair.GetConnection().GetUseTestNet())
	assert.Equal(t, SpotCommissionMaker, pair.GetConnection().GetCommissionMaker())
	assert.Equal(t, SpotCommissionTaker, pair.GetConnection().GetCommissionTaker())

	pair.GetConnection().SetApiKey(FuturesAPIKey)
	pair.GetConnection().SetSecretKey(FuturesAPISecret)
	pair.GetConnection().SetUseTestNet(FuturesUseTestNet)
	pair.GetConnection().SetCommissionMaker(FuturesCommissionMaker)
	pair.GetConnection().SetCommissionTaker(FuturesCommissionTaker)

	pair.SetInitialBalance(3000)
	pair.SetCurrentBalance(4000)
	pair.SetInitialPositionBalance(3000 * LimitOnPosition_2)
	pair.SetCurrentPositionBalance(4000 * LimitOnPosition_2)
	pair.SetMiddlePrice(45000)
	pair.SetStage(StageType_2)
	pair.SetBuyQuantity(BuyQuantity_2)
	pair.SetBuyValue(BuyValue_2)
	pair.SetBuyCommission(BuyCommission_2)
	pair.SetSellQuantity(SellQuantity_2)
	pair.SetSellValue(SellValue_2)
	pair.SetSellCommission(SellCommission_2)
	pair.SetCommission(newCommission)

	assert.Equal(t, FuturesAPIKey, pair.GetConnection().GetAPIKey())
	assert.Equal(t, FuturesAPISecret, pair.GetConnection().GetSecretKey())
	assert.Equal(t, FuturesUseTestNet, pair.GetConnection().GetUseTestNet())
	assert.Equal(t, FuturesCommissionMaker, pair.GetConnection().GetCommissionMaker())
	assert.Equal(t, FuturesCommissionTaker, pair.GetConnection().GetCommissionTaker())

	assert.Equal(t, 3000.0, pair.GetInitialBalance())
	assert.Equal(t, 4000.0, pair.GetCurrentBalance())
	assert.Equal(t, 3000*LimitOnPosition_2, pair.GetInitialPositionBalance())
	assert.Equal(t, 4000*LimitOnPosition_2, pair.GetCurrentPositionBalance())
	assert.Equal(t, 45000.0, pair.GetMiddlePrice())
	assert.Equal(t, StageType_2, pair.GetStage())
	assert.Equal(t, BuyQuantity_2, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_2, pair.GetBuyValue())
	assert.Equal(t, BuyCommission_2, pair.GetBuyCommission())
	assert.Equal(t, SellQuantity_2, pair.GetSellQuantity())
	assert.Equal(t, SellValue_2, pair.GetSellValue())
	assert.Equal(t, SellCommission_2, pair.GetSellCommission())
	assert.Equal(t, newCommission, pair.GetCommission())

	pair.SetBuyData(BuyQuantity_1, BuyValue_1, BuyCommission_1)
	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())
	assert.Equal(t, BuyCommission_1, pair.GetBuyCommission())

	pair.SetSellData(SellQuantity_1, SellValue_1, SellCommission_1)
	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())
	assert.Equal(t, SellCommission_1, pair.GetSellCommission())
}

func TestPairGetter(t *testing.T) {
	pair := pair_1
	assert.Equal(t, "", pair.GetConnection().GetAPIKey())
	assert.Equal(t, "", pair.GetConnection().GetSecretKey())
	assert.Equal(t, false, pair.GetConnection().GetUseTestNet())
	assert.Equal(t, 0.0, pair.GetConnection().GetCommissionMaker())
	assert.Equal(t, 0.0, pair.GetConnection().GetCommissionTaker())
	assert.Equal(t, InitialBalance, pair.GetInitialBalance())
	assert.Equal(t, CurrentBalance, pair.GetCurrentBalance())
	assert.Equal(t, InitialPositionBalance_1, pair.GetInitialPositionBalance())
	assert.Equal(t, CurrentPositionBalance_1, pair.GetCurrentPositionBalance())
	assert.Equal(t, AccountType_1, pair.GetAccountType())
	assert.Equal(t, StrategyType_1, pair.GetStrategy())
	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, Pair_1, pair.GetPair())
	assert.Equal(t, TargetSymbol_1, pair.GetTargetSymbol())
	assert.Equal(t, BaseSymbol_1, pair.GetBaseSymbol())
	assert.Equal(t, MiddlePrice_1, pair.GetMiddlePrice())
	assert.Equal(t, LimitInputIntoPosition_1, pair.GetLimitInputIntoPosition())
	assert.Equal(t, LimitOutputOfPosition_1, pair.GetLimitOutputOfPosition())
	assert.Equal(t, LimitOnPosition_1, pair.GetLimitOnPosition())
	assert.Equal(t, LimitOnTransaction_1, pair.GetLimitOnTransaction())
	assert.Equal(t, BuyDelta_1, pair.GetBuyDelta())
	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())
	assert.Equal(t, BuyCommission_1, pair.GetBuyCommission())
	assert.Equal(t, SellDelta_1, pair.GetSellDelta())
	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())
	assert.Equal(t, SellCommission_1, pair.GetSellCommission())
	assert.Equal(t, Commission, pair.GetCommission())
}

func TestPairChecking(t *testing.T) {
	assert.True(t, pair_1.CheckingPair())
	assert.False(t, pair_2.CheckingPair())
}

func TestConfigGetter(t *testing.T) {
	config := config
	config.Pairs.ReplaceOrInsert(pair_1)
	config.Pairs.ReplaceOrInsert(pair_2)
	assertTest(t, nil, config, &[]pairs_interfaces.Pairs{pair_1, pair_2})
}

func TestConfigSetter(t *testing.T) {
	pairs := []pairs_interfaces.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assertTest(t, err, config, checkingDate)
}

func TestConfigGetPairs(t *testing.T) {
	pairs := []pairs_interfaces.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assertTest(t, err, config, checkingDate)
}
