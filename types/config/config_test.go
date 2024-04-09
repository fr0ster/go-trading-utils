package config_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	config_types "github.com/fr0ster/go-trading-utils/types/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

const (
	APIKey                   = "your_api_key"                     // Ключ API
	APISecret                = "your_api_secret"                  // Секретний ключ API
	UseTestNet               = false                              // Використовувати тестову мережу
	AccountType_1            = pairs_types.SpotAccountType        // Тип акаунта
	StrategyType_1           = pairs_types.HoldingStrategyType    // Тип стратегії
	StageType_1              = pairs_types.InputIntoPositionStage // Стадія стратегії
	Pair_1                   = "BTCUSDT"                          // Пара
	TargetSymbol_1           = "BTC"                              // Котирувальна валюта
	BaseSymbol_1             = "USDT"                             // Базова валюта
	BaseBalance_1            = 2000.0                             // Баланс базової валюти
	LimitInputIntoPosition_1 = 0.01                               // Ліміт на вхід в позицію, відсоток від балансу базової валюти
	LimitOnPosition_1        = 0.50                               // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_1     = 0.10                               // Ліміт на транзакцію, відсоток від ліміту на позицію
	BuyDelta_1               = 0.01                               // Дельта для купівлі
	SellDelta_1              = 0.01                               // Дельта для продажу
	BuyQuantity_1            = 1.0                                // Кількість для купівлі, суммарно по позиції
	SellQuantity_1           = 1.0                                // Кількість для продажу, суммарно по позиції
	BuyValue_1               = 100.0                              // Вартість для купівлі, суммарно по позиції
	SellValue_1              = 100.0                              // Вартість для продажу, суммарно по позиції
	AccountType_2            = pairs_types.USDTFutureType         // Тип акаунта
	StrategyType_2           = pairs_types.TradingStrategyType    // Тип стратегії
	StageType_2              = pairs_types.WorkInPositionStage    // Тип стадії
	Pair_2                   = "ETHUSDT"                          // Пара
	TargetSymbol_2           = "ETH"                              // Котирувальна валюта
	BaseSymbol_2             = "USDT"                             // Базова валюта
	LimitValue_2             = 2000.0                             // Баланс базової валюти
	LimitInputIntoPosition_2 = 0.10                               // Ліміт на вхід в позицію, відсоток від балансу базової валюти
	LimitOnPosition_2        = 0.50                               // Ліміт на позицію, відсоток від балансу базової валюти
	LimitOnTransaction_2     = 0.01                               // Ліміт на транзакцію, відсоток від ліміту на позицію
	BuyDelta_2               = 0.01                               // Дельта для купівлі
	SellDelta_2              = 0.01                               // Дельта для продажу
	BuyQuantity_2            = 1.0                                // Кількість для купівлі, суммарно по позиції
	SellQuantity_2           = 1.0                                // Кількість для продажу, суммарно по позиції
	BuyValue_2               = 100.0                              // Вартість для купівлі, суммарно по позиції
	SellValue_2              = 100.0                              // Вартість для продажу, суммарно по позиції
)

var (
	pair_1 = &pairs_types.Pairs{
		AccountType:            AccountType_1,
		StrategyType:           StrategyType_1,
		StageType:              StageType_1,
		Pair:                   Pair_1,
		TargetSymbol:           TargetSymbol_1,
		BaseSymbol:             BaseSymbol_1,
		LimitInputIntoPosition: LimitInputIntoPosition_1,
		LimitOnPosition:        LimitOnPosition_1,
		LimitOnTransaction:     LimitOnTransaction_1,
		BuyDelta:               BuyDelta_1,
		BuyQuantity:            BuyQuantity_1,
		BuyValue:               BuyValue_1,
		SellDelta:              SellDelta_1,
		SellQuantity:           SellQuantity_1,
		SellValue:              SellValue_1,
	}
	pair_2 = &pairs_types.Pairs{
		AccountType:            AccountType_2,
		StrategyType:           StrategyType_2,
		StageType:              StageType_2,
		Pair:                   Pair_2,
		TargetSymbol:           TargetSymbol_2,
		BaseSymbol:             BaseSymbol_2,
		LimitInputIntoPosition: LimitInputIntoPosition_2,
		LimitOnPosition:        LimitOnPosition_2,
		LimitOnTransaction:     LimitOnTransaction_2,
		BuyDelta:               BuyDelta_2,
		BuyQuantity:            BuyQuantity_2,
		BuyValue:               BuyValue_2,
		SellDelta:              SellDelta_2,
		SellQuantity:           SellQuantity_2,
		SellValue:              SellValue_2,
	}
)

func getTestData() []byte {
	return []byte(`{
		"api_key": "` + APIKey + `",
		"api_secret": "` + APISecret + `",
		"use_test_net": ` + strconv.FormatBool(UseTestNet) + `,
		"pairs": [
			{
				"account_type": "` + string(AccountType_1) + `",
				"strategy_type": "` + string(StrategyType_1) + `",
				"stage_type": "` + string(StageType_1) + `",
				"symbol": "` + Pair_1 + `",
				"target_symbol": "` + TargetSymbol_1 + `",
				"base_symbol": "` + BaseSymbol_1 + `",
				"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitInputIntoPosition_1, 'f', -1, 64)).String() + `,
				"limit_on_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_1, 'f', -1, 64)).String() + `,
				"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_1, 'f', -1, 64)).String() + `,
				"buy_delta": ` + json.Number(strconv.FormatFloat(BuyDelta_1, 'f', -1, 64)).String() + `,
				"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_1, 'f', -1, 64)).String() + `,
				"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_1, 'f', -1, 64)).String() + `,
				"sell_delta": ` + json.Number(strconv.FormatFloat(SellDelta_1, 'f', -1, 64)).String() + `,
				"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_1, 'f', -1, 64)).String() + `,
				"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_1, 'f', -1, 64)).String() + `
			},
			{
				"account_type": "` + string(AccountType_2) + `",
				"strategy_type": "` + string(StrategyType_2) + `",
				"stage_type": "` + string(StageType_2) + `",
				"symbol": "` + Pair_2 + `",
				"target_symbol": "` + TargetSymbol_2 + `",
				"base_symbol": "` + BaseSymbol_2 + `",
				"limit_input_into_position": ` + json.Number(strconv.FormatFloat(LimitValue_2, 'f', -1, 64)).String() + `,
				"limit_in_position": ` + json.Number(strconv.FormatFloat(LimitOnPosition_2, 'f', -1, 64)).String() + `,
				"limit_on_transaction": ` + json.Number(strconv.FormatFloat(LimitOnTransaction_2, 'f', -1, 64)).String() + `,
				"buy_delta": ` + json.Number(strconv.FormatFloat(BuyDelta_2, 'f', -1, 64)).String() + `,
				"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_2, 'f', -1, 64)).String() + `,
				"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_2, 'f', -1, 64)).String() + `,
				"sell_delta": ` + json.Number(strconv.FormatFloat(SellDelta_1, 'f', -1, 64)).String() + `,
				"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_2, 'f', -1, 64)).String() + `,
				"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_2, 'f', -1, 64)).String() + `
				}
			]
		}`)
}

func assertTest(t *testing.T, err error, config config_interfaces.Configuration, checkingDate *[]config_interfaces.Pairs) {
	assert.NoError(t, err)
	assert.Equal(t, APIKey, config.GetAPIKey())
	assert.Equal(t, APISecret, config.GetSecretKey())
	assert.Equal(t, UseTestNet, config.GetUseTestNet())

	assert.Equal(t, (*checkingDate)[0].GetAccountType(), config.GetPair(Pair_1).GetAccountType())
	assert.Equal(t, (*checkingDate)[0].GetStrategy(), config.GetPair(Pair_1).GetStrategy())
	assert.Equal(t, (*checkingDate)[0].GetStage(), config.GetPair(Pair_1).GetStage())
	assert.Equal(t, (*checkingDate)[0].GetPair(), config.GetPair(Pair_1).GetPair())
	assert.Equal(t, (*checkingDate)[0].GetTargetSymbol(), config.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[0].GetBaseSymbol(), config.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[0].GetLimitInputIntoPosition(), config.GetPair(Pair_1).GetLimitInputIntoPosition())
	assert.Equal(t, (*checkingDate)[0].GetLimitOnPosition(), config.GetPair(Pair_1).GetLimitOnPosition())
	assert.Equal(t, (*checkingDate)[0].GetLimitOnTransaction(), config.GetPair(Pair_1).GetLimitOnTransaction())
	assert.Equal(t, (*checkingDate)[0].GetBuyDelta(), config.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[0].GetBuyQuantity(), config.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[0].GetBuyValue(), config.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, (*checkingDate)[0].GetSellDelta(), config.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, (*checkingDate)[0].GetSellQuantity(), config.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[0].GetSellValue(), config.GetPair(Pair_1).GetSellValue())

	assert.Equal(t, (*checkingDate)[1].GetAccountType(), config.GetPair(Pair_2).GetAccountType())
	assert.Equal(t, (*checkingDate)[1].GetStrategy(), config.GetPair(Pair_2).GetStrategy())
	assert.Equal(t, (*checkingDate)[1].GetStage(), config.GetPair(Pair_2).GetStage())
	assert.Equal(t, (*checkingDate)[1].GetPair(), config.GetPair(Pair_2).GetPair())
	assert.Equal(t, (*checkingDate)[1].GetLimitInputIntoPosition(), config.GetPair(Pair_2).GetLimitInputIntoPosition())
	assert.Equal(t, (*checkingDate)[1].GetLimitOnPosition(), config.GetPair(Pair_2).GetLimitOnPosition())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), config.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), config.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetLimitOnTransaction(), config.GetPair(Pair_2).GetLimitOnTransaction())
	assert.Equal(t, (*checkingDate)[1].GetBuyDelta(), config.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[1].GetBuyQuantity(), config.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[1].GetBuyValue(), config.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, (*checkingDate)[1].GetSellDelta(), config.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, (*checkingDate)[1].GetSellQuantity(), config.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[1].GetSellValue(), config.GetPair(Pair_2).GetSellValue())

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
	configFile := config_types.ConfigNew(tmpFile.Name(), 2)

	// Load the config from the file
	err = configFile.Load()
	assert.NoError(t, err)

	// Assert that the loaded config matches the test data
	checkingDate, err := configFile.Configs.GetPairs()
	assertTest(t, err, configFile.GetConfigurations(), checkingDate)
}

func TestConfigFile_Save(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a new ConfigFile instance
	config := &config_types.ConfigFile{
		FilePath: tmpFile.Name(),
		Configs: &config_types.Configs{
			APIKey:     APIKey,
			APISecret:  APISecret,
			UseTestNet: UseTestNet,
			Pairs:      btree.New(2),
		},
	}
	config.Configs.Pairs.ReplaceOrInsert(pair_1)
	config.Configs.Pairs.ReplaceOrInsert(pair_2)

	// Save the config to the file
	err = config.Save()
	assert.NoError(t, err)

	// Read the saved config file
	savedData, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	// Unmarshal the saved data into a ConfigFile struct
	savedConfig := &config_types.Configs{}
	err = json.Unmarshal(savedData, savedConfig)
	assert.NoError(t, err)

	// Assert that the saved config matches the original config
	checkingDate, err := config.GetConfigurations().GetPairs()
	assertTest(t, err, config.GetConfigurations(), checkingDate)
}

// Add more tests for other methods if needed

func TestPairSetter(t *testing.T) {
	pair := &pairs_types.Pairs{
		Pair:        Pair_1,
		BuyQuantity: BuyQuantity_1,
		BuyValue:    BuyValue_1,
	}
	pair.SetStage(StageType_2)
	pair.SetBuyQuantity(BuyQuantity_2)
	pair.SetBuyValue(BuyValue_2)
	pair.SetSellQuantity(SellQuantity_2)
	pair.SetSellValue(SellValue_2)

	assert.Equal(t, StageType_2, pair.GetStage())
	assert.Equal(t, BuyQuantity_2, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_2, pair.GetBuyValue())
	assert.Equal(t, SellQuantity_2, pair.GetSellQuantity())
	assert.Equal(t, SellValue_2, pair.GetSellValue())
}

func TestPairGetter(t *testing.T) {
	pair := pair_1
	assert.Equal(t, AccountType_1, pair.GetAccountType())
	assert.Equal(t, StrategyType_1, pair.GetStrategy())
	assert.Equal(t, StageType_1, pair.GetStage())
	assert.Equal(t, Pair_1, pair.GetPair())
	assert.Equal(t, TargetSymbol_1, pair.GetTargetSymbol())
	assert.Equal(t, BaseSymbol_1, pair.GetBaseSymbol())
	assert.Equal(t, LimitInputIntoPosition_1, pair.GetLimitInputIntoPosition())
	assert.Equal(t, LimitOnPosition_1, pair.GetLimitOnPosition())
	assert.Equal(t, LimitOnTransaction_1, pair.GetLimitOnTransaction())
	assert.Equal(t, BuyDelta_1, pair.GetBuyDelta())
	assert.Equal(t, BuyQuantity_1, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_1, pair.GetBuyValue())
	assert.Equal(t, SellDelta_1, pair.GetSellDelta())
	assert.Equal(t, SellQuantity_1, pair.GetSellQuantity())
	assert.Equal(t, SellValue_1, pair.GetSellValue())
}

func TestConfigGetter(t *testing.T) {
	config := &config_types.Configs{
		APIKey:     APIKey,
		APISecret:  APISecret,
		UseTestNet: UseTestNet,
		Pairs:      btree.New(2),
	}
	config.Pairs.ReplaceOrInsert(pair_1)
	config.Pairs.ReplaceOrInsert(pair_2)

	assertTest(t, nil, config, &[]config_interfaces.Pairs{pair_1, pair_2})
}

func TestConfigSetter(t *testing.T) {
	config := &config_types.Configs{
		APIKey:     APIKey,
		APISecret:  APISecret,
		UseTestNet: UseTestNet,
		Pairs:      btree.New(2),
	}
	pairs := []config_interfaces.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assertTest(t, err, config, checkingDate)
}

func TestConfigGetPairs(t *testing.T) {
	config := &config_types.Configs{
		APIKey:     APIKey,
		APISecret:  APISecret,
		UseTestNet: UseTestNet,
		Pairs:      btree.New(2),
	}
	pairs := []config_interfaces.Pairs{pair_1, pair_2}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assertTest(t, err, config, checkingDate)
}
