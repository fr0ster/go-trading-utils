package config_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
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

	ObservePriceLiquidation    = true // Скасування обмежених ордерів які за лімітом
	ObservePosition            = true // Скасування збитковоі позиції
	ClosePositionOnRestart     = true // Рестарт закритої позиції
	BalancingOfMargin          = true // Балансування маржі
	PercentsToLiquidation      = 0.05 // Відсоток до ліквідації
	PercentToDecreasePosition  = 0.03 // Відсоток для зменшення позиції
	ObserverTimeOutMillisecond = 1000 // Таймаут спостереження
	UsingBreakEvenPrice        = true // Використання ціни без збитків для визначення цін ф'ючерсних ордерів

	MaintainPartiallyFilledOrders = true // Підтримувати частково виконані ордери

	// Для USDT_FUTURE/COIN_FUTURE
	MarginType_1 = pairs_types.CrossMarginType // Кросова маржа
	Leverage_1   = 20                          // Плече 20

	AccountType_1                = pairs_types.SpotAccountType        // Тип акаунта
	StrategyType_1               = pairs_types.HoldingStrategyType    // Тип стратегії
	StageType_1                  = pairs_types.InputIntoPositionStage // Стадія стратегії
	Pair_1                       = "BTCUSDT"                          // Пара
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

	LimitOnPosition_1    = 0.50 // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_1 = 0.10 // Ліміт на транзакцію, відсоток від ліміту на позицію

	UnRealizedProfitLowBound_1 = 0.1 // Нижня межа нереалізованого прибутку
	UnRealizedProfitUpBound_1  = 0.9 // Верхня межа нереалізованого прибутку

	UpBound_1  = 80000.0 // Верхня межа
	LowBound_1 = 40000.0 // Нижня межа

	DeltaPrice_1    = 0.01  // Дельта для купівлі
	DeltaQuantity_1 = 0.1   // Дельта для кількості
	BuyQuantity_1   = 1.0   // Кількість для купівлі, суммарно по позиції
	SellQuantity_1  = 2.0   // Кількість для продажу, суммарно по позиції
	BuyValue_1      = 100.0 // Вартість для купівлі, суммарно по позиції
	SellValue_1     = 200.0 // Вартість для продажу, суммарно по позиції

	CallbackRate_1 = 0.1 // CallbackRate 0.1%

	// Для USDT_FUTURE/COIN_FUTURE
	MarginType_2 = pairs_types.IsolatedMarginType // Ізольована маржа
	Leverage_2   = 10                             // Плече 10

	AccountType_2  = pairs_types.USDTFutureType      // Тип акаунта
	StrategyType_2 = pairs_types.TradingStrategyType // Тип стратегії
	StageType_2    = pairs_types.WorkInPositionStage // Тип стадії
	Pair_2         = "ETHUSDT"                       // Пара

	// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
	// повинно бути меньше ніж LimitOutputOfPosition_2,
	// але це для перевірки CheckingPair
	LimitInputIntoPosition_2 = 0.15
	// Ліміт на вихід з позиції, відсоток від балансу базової валюти,
	// повинно бути більше ніж LimitInputIntoPosition_2,
	// але це для перевірки CheckingPair
	LimitOutputOfPosition_2 = 0.10 // Ліміт на вихід з позиції, відсоток від балансу базової валюти

	LimitOnPosition_2    = 0.50 // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_2 = 0.01 // Ліміт на транзакцію, відсоток від ліміту на позицію

	UnRealizedProfitLowBound_2 = 0.1 // Нижня межа нереалізованого прибутку
	UnRealizedProfitUpBound_2  = 0.9 // Верхня межа нереалізованого прибутку

	UpBound_2  = 14.0 // Верхня межа
	LowBound_2 = 4.0  // Нижня межа

	Delta_Price_2   = 0.01  // Дельта для купівлі
	DeltaQuantity_2 = 0.1   // Дельта для кількості
	BuyQuantity_2   = 1.0   // Кількість для купівлі, суммарно по позиції
	SellQuantity_2  = 1.0   // Кількість для продажу, суммарно по позиції
	BuyValue_2      = 100.0 // Вартість для купівлі, суммарно по позиції
	SellValue_2     = 100.0 // Вартість для продажу, суммарно по позиції

	CallbackRate_2 = 0.5 // CallbackRate 0.5%
)

var (
	DefaultSpotConnection = &connection_types.Connection{
		APIKey:     SpotAPIKey,
		APISecret:  SpotAPISecret,
		UseTestNet: SpotUseTestNet,
	}
	DefaultFuturesConnection = &connection_types.Connection{
		APIKey:     FuturesAPIKey,
		APISecret:  FuturesAPISecret,
		UseTestNet: FuturesUseTestNet,
	}
	config = &config_types.Configs{
		Connection: &connection_types.Connection{
			APIKey:     SpotAPIKey,
			APISecret:  SpotAPISecret,
			UseTestNet: SpotUseTestNet,
		},
		LogLevel:                      InfoLevel,
		ObservePriceLiquidation:       ObservePriceLiquidation,
		ObservePosition:               ObservePosition,
		ClosePositionOnRestart:        ClosePositionOnRestart,
		BalancingOfMargin:             BalancingOfMargin,
		PercentsToStopSettingNewOrder: PercentsToLiquidation,
		PercentToDecreasePosition:     PercentToDecreasePosition,
		ObserverTimeOutMillisecond:    ObserverTimeOutMillisecond,
		Pairs:                         btree.New(2),
	}
	pair_1 = &pairs_types.Pairs{
		AccountType:              AccountType_1,
		StrategyType:             StrategyType_1,
		StageType:                StageType_1,
		Pair:                     Pair_1,
		MarginType:               MarginType_1,
		Leverage:                 Leverage_1,
		LimitInputIntoPosition:   LimitInputIntoPosition_1,
		LimitOutputOfPosition:    LimitOutputOfPosition_1,
		LimitOnPosition:          LimitOnPosition_1,
		LimitOnTransaction:       LimitOnTransaction_1,
		UnRealizedProfitLowBound: UnRealizedProfitLowBound_1,
		UnRealizedProfitUpBound:  UnRealizedProfitUpBound_1,
		UpBound:                  UpBound_1,
		LowBound:                 LowBound_1,
		DeltaPrice:               DeltaPrice_1,
		DeltaQuantity:            DeltaQuantity_1,
		BuyQuantity:              BuyQuantity_1,
		BuyValue:                 BuyValue_1,
		SellQuantity:             SellQuantity_1,
		SellValue:                SellValue_1,
		CallbackRate:             CallbackRate_1,
	}
	pair_2 = &pairs_types.Pairs{
		AccountType:              AccountType_2,
		StrategyType:             StrategyType_2,
		StageType:                StageType_2,
		Pair:                     Pair_2,
		MarginType:               MarginType_2,
		Leverage:                 Leverage_2,
		LimitInputIntoPosition:   LimitInputIntoPosition_2,
		LimitOutputOfPosition:    LimitOutputOfPosition_2,
		LimitOnPosition:          LimitOnPosition_2,
		LimitOnTransaction:       LimitOnTransaction_2,
		UnRealizedProfitLowBound: UnRealizedProfitLowBound_2,
		UnRealizedProfitUpBound:  UnRealizedProfitUpBound_2,
		UpBound:                  UpBound_2,
		LowBound:                 LowBound_2,
		DeltaPrice:               Delta_Price_2,
		DeltaQuantity:            DeltaQuantity_2,
		BuyQuantity:              BuyQuantity_2,
		BuyValue:                 BuyValue_2,
		SellQuantity:             SellQuantity_2,
		SellValue:                SellValue_2,
		CallbackRate:             CallbackRate_2,
	}
)

func getTestData() []byte {
	return []byte(
		`{
			"connection": {
				"api_key": "` + SpotAPIKey + `",
				"api_secret": "` + SpotAPISecret + `",
				"use_test_net": ` + strconv.FormatBool(SpotUseTestNet) + `
			},
			"log_level": "` + InfoLevel.String() + `",
			"observe_price_liquidation": ` + strconv.FormatBool(ObservePriceLiquidation) + `,
			"observe_position": ` + strconv.FormatBool(ObservePosition) + `,
			"close_position_on_restart": ` + strconv.FormatBool(ClosePositionOnRestart) + `,
			"balancing_of_margin": ` + strconv.FormatBool(BalancingOfMargin) + `,
			"percents_to_stop_setting_new_order": ` + json.Number(strconv.FormatFloat(PercentsToLiquidation, 'f', -1, 64)).String() + `,
			"percent_to_decrease_position": ` + json.Number(strconv.FormatFloat(PercentToDecreasePosition, 'f', -1, 64)).String() + `,
			"observer_timeout_millisecond": ` + strconv.Itoa(ObserverTimeOutMillisecond) + `,
			"using_break_even_price": ` + strconv.FormatBool(UsingBreakEvenPrice) + `,
			"pairs": [
				{
					"account_type": "` + string(AccountType_1) + `",
					"strategy_type": "` + string(StrategyType_1) + `",
					"stage_type": "` + string(StageType_1) + `",
					"symbol": "` + Pair_1 + `",
					"margin_type": "` + string(MarginType_1) + `",
					"leverage": ` + strconv.Itoa(Leverage_1) + `,
					"sleeping_time": ` + strconv.Itoa(SleepingTime_1) + `,
					"taking_position_sleeping_time": ` + strconv.Itoa(TakingPositionSleepingTime_1) + `,
					"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitInputIntoPosition_1, 'f', -1, 64)).String() + `,
					"limit_output_of_position": ` + json.Number(strconv.FormatFloat(LimitOutputOfPosition_1, 'f', -1, 64)).String() + `,
					"limit_on_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_1, 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_1, 'f', -1, 64)).String() + `,
					"unrealized_profit_low_bound": ` + json.Number(strconv.FormatFloat(UnRealizedProfitLowBound_1, 'f', -1, 64)).String() + `,
					"unrealized_profit_up_bound": ` + json.Number(strconv.FormatFloat(UnRealizedProfitUpBound_1, 'f', -1, 64)).String() + `,
					"up_bound": ` + json.Number(strconv.FormatFloat(UpBound_1, 'f', -1, 64)).String() + `,
					"low_bound": ` + json.Number(strconv.FormatFloat(LowBound_1, 'f', -1, 64)).String() + `,
					"delta_price": ` + json.Number(strconv.FormatFloat(DeltaPrice_1, 'f', -1, 64)).String() + `,
					"delta_quantity": ` + json.Number(strconv.FormatFloat(DeltaQuantity_1, 'f', -1, 64)).String() + `,
					"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_1, 'f', -1, 64)).String() + `,
					"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_1, 'f', -1, 64)).String() + `,
					"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_1, 'f', -1, 64)).String() + `,
					"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_1, 'f', -1, 64)).String() + `,
					"callback_rate": ` + json.Number(strconv.FormatFloat(CallbackRate_1, 'f', -1, 64)).String() + `
				},
				{
					"account_type": "` + string(AccountType_2) + `",
					"strategy_type": "` + string(StrategyType_2) + `",
					"stage_type": "` + string(StageType_2) + `",
					"symbol": "` + Pair_2 + `",
					"margin_type": "` + string(MarginType_2) + `",
					"leverage": ` + strconv.Itoa(Leverage_2) + `,
					"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitInputIntoPosition_2, 'f', -1, 64)).String() + `,
					"limit_output_of_position": ` + json.Number(strconv.FormatFloat(LimitOutputOfPosition_2, 'f', -1, 64)).String() + `,
					"limit_in_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_2, 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_2, 'f', -1, 64)).String() + `,
					"unrealized_profit_low_bound": ` + json.Number(strconv.FormatFloat(UnRealizedProfitLowBound_2, 'f', -1, 64)).String() + `,
					"unrealized_profit_up_bound": ` + json.Number(strconv.FormatFloat(UnRealizedProfitUpBound_2, 'f', -1, 64)).String() + `,
					"up_bound": ` + json.Number(strconv.FormatFloat(UpBound_2, 'f', -1, 64)).String() + `,
					"low_bound": ` + json.Number(strconv.FormatFloat(LowBound_2, 'f', -1, 64)).String() + `,
					"delta_price": ` + json.Number(strconv.FormatFloat(Delta_Price_2, 'f', -1, 64)).String() + `,
					"buy_delta_quantity": ` + json.Number(strconv.FormatFloat(DeltaQuantity_2, 'f', -1, 64)).String() + `,
					"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_2, 'f', -1, 64)).String() + `,
					"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_2, 'f', -1, 64)).String() + `,
					"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_2, 'f', -1, 64)).String() + `,
					"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_2, 'f', -1, 64)).String() + `,
					"callback_rate": ` + json.Number(strconv.FormatFloat(CallbackRate_2, 'f', -1, 64)).String() + `
				}
			]
		}`)
}

func assertTest(t *testing.T, config config_interfaces.Configuration) {
	checkingDate, err := config.GetPairs()

	assert.NoError(t, err)
	assert.Equal(t, SpotAPIKey, config.GetConnection().GetAPIKey())
	assert.Equal(t, SpotAPISecret, config.GetConnection().GetSecretKey())
	assert.Equal(t, SpotUseTestNet, config.GetConnection().GetUseTestNet())
	assert.Equal(t, InfoLevel, config.GetLogLevel())
	assert.Equal(t, ObservePriceLiquidation, config.GetObservePriceLiquidation())
	assert.Equal(t, ObservePosition, config.GetObservePosition())
	assert.Equal(t, ClosePositionOnRestart, config.GetClosePositionOnRestart())
	assert.Equal(t, BalancingOfMargin, config.GetBalancingOfMargin())
	assert.Equal(t, PercentsToLiquidation, config.GetPercentsToStopSettingNewOrder())
	assert.Equal(t, PercentToDecreasePosition, config.GetPercentToDecreasePosition())
	assert.Equal(t, ObserverTimeOutMillisecond, config.GetObserverTimeOutMillisecond())

	assert.Equal(t, (checkingDate)[0].GetAccountType(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetAccountType())
	assert.Equal(t, (checkingDate)[0].GetStrategy(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetStrategy())
	assert.Equal(t, (checkingDate)[0].GetStage(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetStage())

	assert.Equal(t, (checkingDate)[0].GetPair(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetPair())
	assert.Equal(t, (checkingDate)[0].GetMarginType(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetMarginType())
	assert.Equal(t, (checkingDate)[0].GetLeverage(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLeverage())

	assert.Equal(t, (checkingDate)[0].GetMiddlePrice(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetMiddlePrice())
	assert.Equal(t, (checkingDate)[0].GetLimitInputIntoPosition(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitInputIntoPosition())
	assert.Equal(t, (checkingDate)[0].GetLimitOutputOfPosition(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitOutputOfPosition())
	assert.Equal(t, (checkingDate)[0].GetLimitOnPosition(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitOnPosition())
	assert.Equal(t, (checkingDate)[0].GetLimitOnTransaction(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitOnTransaction())

	assert.Equal(t, (checkingDate)[0].GetUnRealizedProfitLowBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetUnRealizedProfitLowBound())
	assert.Equal(t, (checkingDate)[0].GetUnRealizedProfitUpBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetUnRealizedProfitUpBound())

	assert.Equal(t, (checkingDate)[0].GetUpBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetUpBound())
	assert.Equal(t, (checkingDate)[0].GetLowBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLowBound())

	assert.Equal(t, (checkingDate)[0].GetBuyQuantity(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetBuyQuantity())
	assert.Equal(t, (checkingDate)[0].GetBuyValue(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetBuyValue())

	assert.Equal(t, (checkingDate)[0].GetSellQuantity(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetSellQuantity())
	assert.Equal(t, (checkingDate)[0].GetSellValue(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetSellValue())

	assert.Equal(t, (checkingDate)[0].GetCallbackRate(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetCallbackRate())

	assert.Equal(t, (checkingDate)[1].GetAccountType(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetAccountType())
	assert.Equal(t, (checkingDate)[1].GetStrategy(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetStrategy())
	assert.Equal(t, (checkingDate)[1].GetStage(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetStage())

	assert.Equal(t, (checkingDate)[1].GetPair(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetPair())
	assert.Equal(t, (checkingDate)[1].GetMarginType(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetMarginType())
	assert.Equal(t, (checkingDate)[1].GetLeverage(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLeverage())

	assert.Equal(t, (checkingDate)[1].GetMiddlePrice(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetMiddlePrice())
	assert.Equal(t, (checkingDate)[1].GetLimitInputIntoPosition(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitInputIntoPosition())
	assert.Equal(t, (checkingDate)[1].GetLimitOutputOfPosition(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitOutputOfPosition())
	assert.Equal(t, (checkingDate)[1].GetLimitOnPosition(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitOnPosition())
	assert.Equal(t, (checkingDate)[1].GetLimitOnTransaction(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitOnTransaction())

	assert.Equal(t, (checkingDate)[1].GetUnRealizedProfitLowBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetUnRealizedProfitLowBound())
	assert.Equal(t, (checkingDate)[1].GetUnRealizedProfitUpBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetUnRealizedProfitUpBound())

	assert.Equal(t, (checkingDate)[1].GetUpBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetUpBound())
	assert.Equal(t, (checkingDate)[1].GetLowBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLowBound())

	assert.Equal(t, (checkingDate)[1].GetBuyQuantity(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetBuyQuantity())
	assert.Equal(t, (checkingDate)[1].GetBuyValue(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetBuyValue())

	assert.Equal(t, (checkingDate)[1].GetSellQuantity(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetSellQuantity())
	assert.Equal(t, (checkingDate)[1].GetSellValue(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetSellValue())

	assert.Equal(t, (checkingDate)[1].GetCallbackRate(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetCallbackRate())

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
	configFile := config_types.NewConfigFile(tmpFile.Name())
	configFile.SetConfigurations(config)

	// Load the config from the file
	err = configFile.Load()
	assert.NoError(t, err)

	// Assert that the loaded config matches the test data
	assertTest(t, configFile.GetConfigurations())
}

func TestConfigFile_Save(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a new ConfigFile instance
	config_file := config_types.NewConfigFile(tmpFile.Name())
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
	assertTest(t, config_file.GetConfigurations())
}

func TestConfigFile_Change(t *testing.T) {
	// Create a temporary config file for testing
	pair := *pair_1
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a new ConfigFile instance
	config_file := config_types.NewConfigFile(tmpFile.Name())
	config_file.SetConfigurations(config)
	config_file.GetConfigurations().SetPair(&pair)
	config_file.GetConfigurations().SetPair(pair_2)

	config_file_test := config_types.NewConfigFile(tmpFile.Name())
	config_file_test.SetConfigurations(config)
	config_file_test.GetConfigurations().SetPair(&pair)
	config_file_test.GetConfigurations().SetPair(pair_2)

	// Save the config to the file
	err = config_file.Save()
	assert.NoError(t, err)

	// Read the saved config file
	config_file_test.Load()
}

// Add more tests for other methods if needed

func TestPairSetter(t *testing.T) {
	pair := &pairs_types.Pairs{
		Pair:            Pair_1,
		BuyQuantity:     BuyQuantity_1,
		BuyValue:        BuyValue_1,
		LimitOnPosition: LimitOnPosition_1,
	}

	pair.SetStage(StageType_1)
	pair.SetBuyQuantity(BuyQuantity_1)
	pair.SetBuyValue(BuyValue_1)
	pair.SetSellQuantity(SellQuantity_1)
	pair.SetSellValue(SellValue_1)

	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())
	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())

	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())

	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())
}

func TestPairGetter(t *testing.T) {
	pair := pair_1
	assert.Equal(t, AccountType_1, pair.GetAccountType())
	assert.Equal(t, StrategyType_1, pair.GetStrategy())
	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, Pair_1, pair.GetPair())
	assert.Equal(t, LimitInputIntoPosition_1, pair.GetLimitInputIntoPosition())
	assert.Equal(t, LimitOutputOfPosition_1, pair.GetLimitOutputOfPosition())
	assert.Equal(t, LimitOnPosition_1, pair.GetLimitOnPosition())
	assert.Equal(t, LimitOnTransaction_1, pair.GetLimitOnTransaction())
	assert.Equal(t, UpBound_1, pair.GetUpBound())
	assert.Equal(t, LowBound_1, pair.GetLowBound())
	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())
	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())
}

func TestPairChecking(t *testing.T) {
	assert.True(t, pair_1.CheckingPair())
	assert.False(t, pair_2.CheckingPair())
}

func TestConfigGetter(t *testing.T) {
	config := config
	config.Pairs.ReplaceOrInsert(pair_1)
	config.Pairs.ReplaceOrInsert(pair_2)
	assertTest(t, config)
}

func TestConfigSetter(t *testing.T) {
	pairs := []*pairs_types.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	assertTest(t, config)
}

func TestConfigGetPairs(t *testing.T) {
	pairs := []*pairs_types.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	assertTest(t, config)
}
