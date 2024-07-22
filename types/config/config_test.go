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
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
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

	ObservePriceLiquidation                                 = true // Скасування обмежених ордерів які за лімітом
	ObservePosition                                         = true // Скасування збитковоі позиції
	ClosePositionOnRestart                                  = true // Рестарт закритої позиції
	BalancingOfMargin                                       = true // Балансування маржі
	PercentsToLiquidation      items_types.PricePercentType = 0.05 // Відсоток до ліквідації
	PercentToDecreasePosition  items_types.PricePercentType = 0.03 // Відсоток для зменшення позиції
	ObserverTimeOutMillisecond                              = 1000 // Таймаут спостереження
	UsingBreakEvenPrice                                     = true // Використання ціни без збитків для визначення цін ф'ючерсних ордерів

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

	LimitOnPosition_1    items_types.ValueType        = 1000.0 // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_1 items_types.ValuePercentType = 10.0   // Ліміт на транзакцію, відсоток від ліміту на позицію

	UpAndLowBoundPercent_1 items_types.PricePercentType = 10.0 // Верхня межа відсоток

	MinSteps_1 = 10 // Мінімальна кількість кроків

	DeltaPrice_1    items_types.PricePercentType    = 1.0   // Дельта для купівлі
	DeltaQuantity_1 items_types.QuantityPercentType = 10.0  // Дельта для кількості
	Value_1         items_types.ValueType           = 100.0 // Вартість для позиції

	CallbackRate_1 items_types.PricePercentType = 0.1 // CallbackRate 0.1%

	PercentToTarget_1 items_types.PricePercentType = 10.0 // Відсоток до цілі
	DepthsN_1         int                          = 50   // Глибина

	// Для USDT_FUTURE/COIN_FUTURE
	MarginType_2 = pairs_types.IsolatedMarginType // Ізольована маржа
	Leverage_2   = 10                             // Плече 10

	AccountType_2  = pairs_types.USDTFutureType      // Тип акаунта
	StrategyType_2 = pairs_types.TradingStrategyType // Тип стратегії
	StageType_2    = pairs_types.WorkInPositionStage // Тип стадії
	Pair_2         = "ETHUSDT"                       // Пара

	LimitOnPosition_2    items_types.ValueType        = 2000.0 // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_2 items_types.ValuePercentType = 1.0    // Ліміт на транзакцію, відсоток від ліміту на позицію

	UpAndLowBoundPercent_2 items_types.PricePercentType = 10.0 // Верхня межа відсоток

	MinSteps_2 = 10 // Мінімальна кількість кроків

	Delta_Price_2   items_types.PricePercentType    = 1.0   // Дельта для купівлі
	DeltaQuantity_2 items_types.QuantityPercentType = 10.0  // Дельта для кількості
	Value_2         items_types.ValueType           = 100.0 // Вартість для позиції

	CallbackRate_2 items_types.PricePercentType = 0.5 // CallbackRate 0.5%

	PercentToTarget_2 items_types.PricePercentType = 10  // Відсоток до цілі
	DepthsN_2         int                          = 500 // Глибина
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
		AccountType:        AccountType_1,
		StrategyType:       StrategyType_1,
		StageType:          StageType_1,
		Pair:               Pair_1,
		MarginType:         MarginType_1,
		Leverage:           Leverage_1,
		LimitOnPosition:    LimitOnPosition_1,
		LimitOnTransaction: LimitOnTransaction_1,
		UpAndLowBound:      UpAndLowBoundPercent_1,
		MinSteps:           MinSteps_1,
		DeltaPrice:         DeltaPrice_1,
		DeltaQuantity:      DeltaQuantity_1,
		Value:              Value_1,
		CallbackRate:       CallbackRate_1,
		PercentToTarget:    PercentToTarget_1,
		DepthsN:            DepthsN_1,
	}
	pair_2 = &pairs_types.Pairs{
		AccountType:        AccountType_2,
		StrategyType:       StrategyType_2,
		StageType:          StageType_2,
		Pair:               Pair_2,
		MarginType:         MarginType_2,
		Leverage:           Leverage_2,
		LimitOnPosition:    LimitOnPosition_2,
		LimitOnTransaction: LimitOnTransaction_2,
		UpAndLowBound:      UpAndLowBoundPercent_2,
		MinSteps:           MinSteps_2,
		DeltaPrice:         Delta_Price_2,
		DeltaQuantity:      DeltaQuantity_2,
		Value:              Value_2,
		CallbackRate:       CallbackRate_2,
		PercentToTarget:    PercentToTarget_2,
		DepthsN:            DepthsN_2,
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
			"percents_to_stop_setting_new_order": ` + json.Number(strconv.FormatFloat(float64(PercentsToLiquidation), 'f', -1, 64)).String() + `,
			"percent_to_decrease_position": ` + json.Number(strconv.FormatFloat(float64(PercentToDecreasePosition), 'f', -1, 64)).String() + `,
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
					"limit_on_position": ` + json.Number(strconv.FormatFloat(float64(LimitOnPosition_1), 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(float64(LimitOnTransaction_1), 'f', -1, 64)).String() + `,
					"up_and_low_bound": ` + json.Number(strconv.FormatFloat(float64(UpAndLowBoundPercent_1), 'f', -1, 64)).String() + `,
					"min_steps": ` + strconv.Itoa(MinSteps_1) + `,
					"delta_price": ` + json.Number(strconv.FormatFloat(float64(DeltaPrice_1), 'f', -1, 64)).String() + `,
					"delta_quantity": ` + json.Number(strconv.FormatFloat(float64(DeltaQuantity_1), 'f', -1, 64)).String() + `,
					"value": ` + json.Number(strconv.FormatFloat(float64(Value_1), 'f', -1, 64)).String() + `,
					"callback_rate": ` + json.Number(strconv.FormatFloat(float64(CallbackRate_1), 'f', -1, 64)).String() + `,
					"percent_to_target": ` + strconv.Itoa(int(PercentToTarget_1)) + `,
					"depths_n": ` + strconv.Itoa(DepthsN_1) + `
				},
				{
					"account_type": "` + string(AccountType_2) + `",
					"strategy_type": "` + string(StrategyType_2) + `",
					"stage_type": "` + string(StageType_2) + `",
					"symbol": "` + Pair_2 + `",
					"margin_type": "` + string(MarginType_2) + `",
					"leverage": ` + strconv.Itoa(Leverage_2) + `,
					"limit_in_position": ` + json.Number(strconv.FormatFloat(float64(LimitOnPosition_2), 'f', -1, 64)).String() + `,
					"limit_on_transaction": ` + json.Number(strconv.FormatFloat(float64(LimitOnTransaction_2), 'f', -1, 64)).String() + `,
					"up_and_low_bound": ` + json.Number(strconv.FormatFloat(float64(UpAndLowBoundPercent_2), 'f', -1, 64)).String() + `,
					"min_steps": ` + strconv.Itoa(MinSteps_2) + `,
					"delta_price": ` + json.Number(strconv.FormatFloat(float64(Delta_Price_2), 'f', -1, 64)).String() + `,
					"buy_delta_quantity": ` + json.Number(strconv.FormatFloat(float64(DeltaQuantity_2), 'f', -1, 64)).String() + `,
					"value": ` + json.Number(strconv.FormatFloat(float64(Value_2), 'f', -1, 64)).String() + `,
					"callback_rate": ` + json.Number(strconv.FormatFloat(float64(CallbackRate_2), 'f', -1, 64)).String() + `,
					"percent_to_target": ` + strconv.Itoa(int(PercentToTarget_2)) + `,
					"depths_n": ` + strconv.Itoa(DepthsN_2) + `
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

	assert.Equal(t, (checkingDate)[0].GetLimitOnPosition(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitOnPosition())
	assert.Equal(t, (checkingDate)[0].GetLimitOnTransaction(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLimitOnTransaction())

	assert.Equal(t, (checkingDate)[0].GetUpBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetUpBound())
	assert.Equal(t, (checkingDate)[0].GetLowBound(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetLowBound())
	assert.Equal(t, (checkingDate)[0].GetMinSteps(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetMinSteps())

	assert.Equal(t, (checkingDate)[0].GetValue(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetValue())

	assert.Equal(t, (checkingDate)[0].GetCallbackRate(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetCallbackRate())

	assert.Equal(t, (checkingDate)[0].GetPercentToTarget(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetPercentToTarget())
	assert.Equal(t, (checkingDate)[0].GetDepthsN(), config.GetPair(AccountType_1, StrategyType_1, StageType_1, Pair_1).GetDepthsN())

	assert.Equal(t, (checkingDate)[1].GetAccountType(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetAccountType())
	assert.Equal(t, (checkingDate)[1].GetStrategy(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetStrategy())
	assert.Equal(t, (checkingDate)[1].GetStage(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetStage())

	assert.Equal(t, (checkingDate)[1].GetPair(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetPair())
	assert.Equal(t, (checkingDate)[1].GetMarginType(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetMarginType())
	assert.Equal(t, (checkingDate)[1].GetLeverage(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLeverage())

	assert.Equal(t, (checkingDate)[1].GetLimitOnPosition(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitOnPosition())
	assert.Equal(t, (checkingDate)[1].GetLimitOnTransaction(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLimitOnTransaction())

	assert.Equal(t, (checkingDate)[1].GetUpBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetUpBound())
	assert.Equal(t, (checkingDate)[1].GetLowBound(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetLowBound())
	assert.Equal(t, (checkingDate)[1].GetMinSteps(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetMinSteps())

	assert.Equal(t, (checkingDate)[1].GetValue(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetValue())

	assert.Equal(t, (checkingDate)[1].GetCallbackRate(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetCallbackRate())

	assert.Equal(t, (checkingDate)[1].GetPercentToTarget(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetPercentToTarget())
	assert.Equal(t, (checkingDate)[1].GetDepthsN(), config.GetPair(AccountType_2, StrategyType_2, StageType_2, Pair_2).GetDepthsN())
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
		LimitOnPosition: LimitOnPosition_1,
	}

	pair.SetStage(StageType_1)
	pair.SetValue(Value_1)

	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, Value_1, pair.GetValue())
}

func TestPairGetter(t *testing.T) {
	pair := pair_1
	assert.Equal(t, AccountType_1, pair.GetAccountType())
	assert.Equal(t, StrategyType_1, pair.GetStrategy())
	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, Pair_1, pair.GetPair())
	assert.Equal(t, LimitOnPosition_1, pair.GetLimitOnPosition())
	assert.Equal(t, LimitOnTransaction_1, pair.GetLimitOnTransaction())
	assert.Equal(t, UpAndLowBoundPercent_1, pair.GetUpBound())
	assert.Equal(t, Value_1, pair.GetValue())
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
