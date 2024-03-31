package config_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

const (
	APIKey         = "your_api_key"
	APISecret      = "your_api_secret"
	UseTestNet     = false
	Pair_1         = "BTCUSDT"
	TargetSymbol_1 = "BTC"
	BaseSymbol_1   = "USDT"
	Limit_1        = 10.0
	Quantity_1     = 1.0
	Value_1        = 100.0
	Pair_2         = "ETHUSDT"
	TargetSymbol_2 = "ETH"
	BaseSymbol_2   = "USDT"
	Limit_2        = 10.0
	Quantity_2     = 1.0
	Value_2        = 100.0
)

func getTestData() []byte {
	return []byte(`{
		"api_key": "` + APIKey + `",
		"api_secret": "` + APISecret + `",
		"use_test_net": ` + strconv.FormatBool(UseTestNet) + `,
		"pairs": [
			{
				"symbol": "` + Pair_1 + `",
				"target_symbol": "` + TargetSymbol_1 + `",
				"base_symbol": "` + BaseSymbol_1 + `",
				"limit": ` + json.Number(strconv.FormatFloat(Limit_1, 'f', -1, 64)).String() + `,
				"quantity": ` + json.Number(strconv.FormatFloat(Quantity_1, 'f', -1, 64)).String() + `,
				"value": ` + json.Number(strconv.FormatFloat(Value_1, 'f', -1, 64)).String() + `
			},
			{
				"symbol": "` + Pair_2 + `",
					"target_symbol": "` + TargetSymbol_2 + `",
					"base_symbol": "` + BaseSymbol_2 + `",
					"limit": ` + json.Number(strconv.FormatFloat(Limit_2, 'f', -1, 64)).String() + `,
					"quantity": ` + json.Number(strconv.FormatFloat(Quantity_2, 'f', -1, 64)).String() + `,
					"value": ` + json.Number(strconv.FormatFloat(Value_2, 'f', -1, 64)).String() + `
				}
			]
		}`)
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
	assert.NoError(t, err)
	assert.Equal(t, APIKey, configFile.Configs.APIKey)
	assert.Equal(t, APISecret, configFile.Configs.APISecret)
	assert.Equal(t, UseTestNet, configFile.Configs.UseTestNet)
	assert.Equal(t, Pair_1, checkingDate[0].GetPair())
	assert.Equal(t, TargetSymbol_1, checkingDate[0].GetTargetSymbol())
	assert.Equal(t, BaseSymbol_1, checkingDate[0].GetBaseSymbol())
	assert.Equal(t, Limit_1, checkingDate[0].GetLimit())
	assert.Equal(t, Quantity_1, checkingDate[0].GetQuantity())
	assert.Equal(t, Value_1, checkingDate[0].GetValue())
	assert.Equal(t, Pair_2, checkingDate[1].GetPair())
	assert.Equal(t, TargetSymbol_2, checkingDate[1].GetTargetSymbol())
	assert.Equal(t, BaseSymbol_2, checkingDate[1].GetBaseSymbol())
	assert.Equal(t, Limit_2, checkingDate[1].GetLimit())
	assert.Equal(t, Quantity_2, checkingDate[1].GetQuantity())
	assert.Equal(t, Value_2, checkingDate[1].GetValue())
}

func TestConfigFile_Save(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a new ConfigFile instance
	config := &config_types.ConfigFile{
		FilePath: tmpFile.Name(),
		Configs: config_types.Configs{
			APIKey:     APIKey,
			APISecret:  APISecret,
			UseTestNet: UseTestNet,
			Pairs:      btree.New(2),
		},
	}
	config.Configs.Pairs.ReplaceOrInsert(&config_types.Pairs{
		Pair:         Pair_1,
		TargetSymbol: TargetSymbol_1,
		BaseSymbol:   BaseSymbol_1,
		Limit:        Limit_1,
		Quantity:     Quantity_1,
		Value:        Value_1,
	})
	config.Configs.Pairs.ReplaceOrInsert(&config_types.Pairs{
		Pair:         Pair_2,
		TargetSymbol: TargetSymbol_2,
		BaseSymbol:   BaseSymbol_2,
		Limit:        Limit_2,
		Quantity:     Quantity_2,
		Value:        Value_2,
	})

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
	assert.NoError(t, err)
	assert.Equal(t, config.GetConfigurations().GetAPIKey(), savedConfig.GetAPIKey())
	assert.Equal(t, config.GetConfigurations().GetSecretKey(), savedConfig.GetSecretKey())
	assert.Equal(t, config.GetConfigurations().GetUseTestNet(), savedConfig.GetUseTestNet())
	assert.Equal(t, checkingDate[0].GetPair(), savedConfig.GetPair(Pair_1).GetPair())
	assert.Equal(t, checkingDate[0].GetTargetSymbol(), savedConfig.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, checkingDate[0].GetBaseSymbol(), savedConfig.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, checkingDate[0].GetLimit(), savedConfig.GetPair(Pair_1).GetLimit())
	assert.Equal(t, checkingDate[0].GetQuantity(), savedConfig.GetPair(Pair_1).GetQuantity())
	assert.Equal(t, checkingDate[0].GetValue(), savedConfig.GetPair(Pair_1).GetValue())
	assert.Equal(t, checkingDate[1].GetPair(), savedConfig.GetPair(Pair_2).GetPair())
	assert.Equal(t, checkingDate[1].GetTargetSymbol(), savedConfig.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, checkingDate[1].GetBaseSymbol(), savedConfig.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, checkingDate[1].GetLimit(), savedConfig.GetPair(Pair_2).GetLimit())
	assert.Equal(t, checkingDate[1].GetQuantity(), savedConfig.GetPair(Pair_2).GetQuantity())
	assert.Equal(t, checkingDate[1].GetValue(), savedConfig.GetPair(Pair_2).GetValue())
}

// Add more tests for other methods if needed
