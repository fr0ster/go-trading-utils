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
	APIKey         = "your_api_key"
	APISecret      = "your_api_secret"
	UseTestNet     = false
	AccountType_1  = pairs_types.SpotAccountType
	Pair_1         = "BTCUSDT"
	TargetSymbol_1 = "BTC"
	BaseSymbol_1   = "USDT"
	Limit_1        = 0.01
	BuyDelta_1     = 0.01
	SellDelta_1    = 0.01
	BuyQuantity_1  = 1.0
	SellQuantity_1 = 1.0
	BuyValue_1     = 100.0
	SellValue_1    = 100.0
	AccountType_2  = pairs_types.USDTFutureType
	Pair_2         = "ETHUSDT"
	TargetSymbol_2 = "ETH"
	BaseSymbol_2   = "USDT"
	Limit_2        = 0.01
	BuyDelta_2     = 0.01
	SellDelta_2    = 0.01
	BuyQuantity_2  = 1.0
	SellQuantity_2 = 1.0
	BuyValue_2     = 100.0
	SellValue_2    = 100.0
)

func getTestData() []byte {
	return []byte(`{
		"api_key": "` + APIKey + `",
		"api_secret": "` + APISecret + `",
		"use_test_net": ` + strconv.FormatBool(UseTestNet) + `,
		"pairs": [
			{
				"account_type": "` + string(AccountType_1) + `",
				"symbol": "` + Pair_1 + `",
				"target_symbol": "` + TargetSymbol_1 + `",
				"base_symbol": "` + BaseSymbol_1 + `",
				"limit": ` + json.Number(strconv.FormatFloat(Limit_1, 'f', -1, 64)).String() + `,
				"buy_delta": ` + json.Number(strconv.FormatFloat(BuyDelta_1, 'f', -1, 64)).String() + `,
				"buy_quantity": ` + json.Number(strconv.FormatFloat(BuyQuantity_1, 'f', -1, 64)).String() + `,
				"buy_value": ` + json.Number(strconv.FormatFloat(BuyValue_1, 'f', -1, 64)).String() + `,
				"sell_delta": ` + json.Number(strconv.FormatFloat(SellDelta_1, 'f', -1, 64)).String() + `,
				"sell_quantity": ` + json.Number(strconv.FormatFloat(SellQuantity_1, 'f', -1, 64)).String() + `,
				"sell_value": ` + json.Number(strconv.FormatFloat(SellValue_1, 'f', -1, 64)).String() + `
			},
			{
				"account_type": "` + string(AccountType_2) + `",
				"symbol": "` + Pair_2 + `",
				"target_symbol": "` + TargetSymbol_2 + `",
				"base_symbol": "` + BaseSymbol_2 + `",
				"limit": ` + json.Number(strconv.FormatFloat(Limit_2, 'f', -1, 64)).String() + `,
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

	assert.Equal(t, Pair_1, (*checkingDate)[0].GetPair())
	assert.Equal(t, TargetSymbol_1, (*checkingDate)[0].GetTargetSymbol())
	assert.Equal(t, BaseSymbol_1, (*checkingDate)[0].GetBaseSymbol())
	assert.Equal(t, Limit_1, (*checkingDate)[0].GetLimit())
	assert.Equal(t, BuyDelta_1, (*checkingDate)[0].GetBuyDelta())
	assert.Equal(t, SellDelta_1, (*checkingDate)[0].GetSellDelta())
	assert.Equal(t, BuyQuantity_1, (*checkingDate)[0].GetBuyQuantity())
	assert.Equal(t, BuyValue_1, (*checkingDate)[0].GetBuyValue())
	assert.Equal(t, SellQuantity_1, (*checkingDate)[0].GetSellQuantity())
	assert.Equal(t, SellValue_1, (*checkingDate)[0].GetSellValue())

	assert.Equal(t, Pair_2, (*checkingDate)[1].GetPair())
	assert.Equal(t, TargetSymbol_2, (*checkingDate)[1].GetTargetSymbol())
	assert.Equal(t, BaseSymbol_2, (*checkingDate)[1].GetBaseSymbol())
	assert.Equal(t, Limit_2, (*checkingDate)[1].GetLimit())
	assert.Equal(t, BuyDelta_2, (*checkingDate)[1].GetBuyDelta())
	assert.Equal(t, SellDelta_1, (*checkingDate)[1].GetSellDelta())
	assert.Equal(t, BuyQuantity_2, (*checkingDate)[1].GetBuyQuantity())
	assert.Equal(t, BuyValue_2, (*checkingDate)[1].GetBuyValue())
	assert.Equal(t, SellQuantity_2, (*checkingDate)[1].GetSellQuantity())
	assert.Equal(t, SellValue_2, (*checkingDate)[1].GetSellValue())
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
	config.Configs.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
		Pair:         Pair_1,
		TargetSymbol: TargetSymbol_1,
		BaseSymbol:   BaseSymbol_1,
		Limit:        Limit_1,
		BuyDelta:     BuyDelta_1,
		BuyQuantity:  BuyQuantity_1,
		BuyValue:     BuyValue_1,
		SellDelta:    SellDelta_1,
		SellQuantity: SellQuantity_1,
		SellValue:    SellValue_1,
	})
	config.Configs.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
		Pair:         Pair_2,
		TargetSymbol: TargetSymbol_2,
		BaseSymbol:   BaseSymbol_2,
		Limit:        Limit_2,
		BuyDelta:     BuyDelta_2,
		BuyQuantity:  BuyQuantity_2,
		BuyValue:     BuyValue_2,
		SellDelta:    SellDelta_2,
		SellQuantity: SellQuantity_2,
		SellValue:    SellValue_2,
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

	assert.Equal(t, (*checkingDate)[0].GetPair(), savedConfig.GetPair(Pair_1).GetPair())
	assert.Equal(t, (*checkingDate)[0].GetTargetSymbol(), savedConfig.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[0].GetBaseSymbol(), savedConfig.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[0].GetLimit(), savedConfig.GetPair(Pair_1).GetLimit())
	assert.Equal(t, (*checkingDate)[0].GetBuyDelta(), savedConfig.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[0].GetBuyQuantity(), savedConfig.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[0].GetBuyValue(), savedConfig.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, (*checkingDate)[0].GetSellDelta(), savedConfig.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, (*checkingDate)[0].GetSellQuantity(), savedConfig.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[0].GetSellValue(), savedConfig.GetPair(Pair_1).GetSellValue())

	assert.Equal(t, (*checkingDate)[1].GetPair(), savedConfig.GetPair(Pair_2).GetPair())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), savedConfig.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), savedConfig.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetLimit(), savedConfig.GetPair(Pair_2).GetLimit())
	assert.Equal(t, (*checkingDate)[1].GetBuyDelta(), savedConfig.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[1].GetBuyQuantity(), savedConfig.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[1].GetBuyValue(), savedConfig.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, (*checkingDate)[1].GetSellDelta(), savedConfig.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, (*checkingDate)[1].GetSellQuantity(), savedConfig.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[1].GetSellValue(), savedConfig.GetPair(Pair_2).GetSellValue())
}

// Add more tests for other methods if needed

func TestPairSetter(t *testing.T) {
	pair := &pairs_types.Pairs{
		Pair:        Pair_1,
		BuyQuantity: BuyQuantity_1,
		BuyValue:    BuyValue_1,
	}
	pair.SetBuyQuantity(BuyQuantity_2)
	pair.SetBuyValue(BuyValue_2)
	pair.SetSellQuantity(SellQuantity_2)
	pair.SetSellValue(SellValue_2)

	assert.Equal(t, BuyQuantity_2, pair.GetBuyQuantity())
	assert.Equal(t, BuyValue_2, pair.GetBuyValue())
	assert.Equal(t, SellQuantity_2, pair.GetSellQuantity())
	assert.Equal(t, SellValue_2, pair.GetSellValue())
}

func TestPairGetter(t *testing.T) {
	pair := &pairs_types.Pairs{
		AccountType:  AccountType_1,
		Pair:         Pair_1,
		TargetSymbol: TargetSymbol_1,
		BaseSymbol:   BaseSymbol_1,
		Limit:        Limit_1,
		BuyDelta:     BuyDelta_1,
		BuyQuantity:  BuyQuantity_1,
		BuyValue:     BuyValue_1,
		SellDelta:    SellDelta_1,
		SellQuantity: SellQuantity_1,
		SellValue:    SellValue_1,
	}
	assert.Equal(t, AccountType_1, pair.GetAccountType())
	assert.Equal(t, Pair_1, pair.GetPair())
	assert.Equal(t, TargetSymbol_1, pair.GetTargetSymbol())
	assert.Equal(t, BaseSymbol_1, pair.GetBaseSymbol())
	assert.Equal(t, Limit_1, pair.GetLimit())
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
	config.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:  AccountType_1,
		Pair:         Pair_1,
		Limit:        Limit_1,
		BuyDelta:     BuyDelta_1,
		BuyQuantity:  BuyQuantity_1,
		BuyValue:     BuyValue_1,
		SellDelta:    SellDelta_1,
		SellQuantity: SellQuantity_1,
		SellValue:    SellValue_1,
	})
	config.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
		AccountType:  AccountType_2,
		Pair:         Pair_2,
		Limit:        Limit_2,
		BuyDelta:     BuyDelta_2,
		BuyQuantity:  BuyQuantity_2,
		BuyValue:     BuyValue_2,
		SellDelta:    SellDelta_2,
		SellQuantity: SellQuantity_2,
		SellValue:    SellValue_2,
	})

	assert.Equal(t, APIKey, config.GetAPIKey())
	assert.Equal(t, APISecret, config.GetSecretKey())
	assert.Equal(t, UseTestNet, config.GetUseTestNet())

	assert.Equal(t, AccountType_1, config.GetPair(Pair_1).GetAccountType())
	assert.Equal(t, Pair_1, config.GetPair(Pair_1).GetPair())
	assert.Equal(t, Limit_1, config.GetPair(Pair_1).GetLimit())
	assert.Equal(t, BuyDelta_1, config.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, BuyQuantity_1, config.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, BuyValue_1, config.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, SellDelta_1, config.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, SellQuantity_1, config.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, SellValue_1, config.GetPair(Pair_1).GetSellValue())

	assert.Equal(t, AccountType_2, config.GetPair(Pair_2).GetAccountType())
	assert.Equal(t, Pair_2, config.GetPair(Pair_2).GetPair())
	assert.Equal(t, Limit_2, config.GetPair(Pair_2).GetLimit())
	assert.Equal(t, BuyDelta_2, config.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, BuyQuantity_2, config.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, BuyValue_2, config.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, SellDelta_2, config.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, SellQuantity_2, config.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, SellValue_2, config.GetPair(Pair_2).GetSellValue())
}

func TestConfigSetter(t *testing.T) {
	config := &config_types.Configs{
		APIKey:     APIKey,
		APISecret:  APISecret,
		UseTestNet: UseTestNet,
		Pairs:      btree.New(2),
	}
	pairs := []config_interfaces.Pairs{
		&pairs_types.Pairs{
			AccountType:  AccountType_1,
			Pair:         Pair_1,
			TargetSymbol: TargetSymbol_1,
			BaseSymbol:   BaseSymbol_1,
			Limit:        Limit_1,
			BuyDelta:     BuyDelta_1,
			BuyQuantity:  BuyQuantity_1,
			BuyValue:     BuyValue_1,
			SellDelta:    SellDelta_1,
			SellQuantity: SellQuantity_1,
			SellValue:    SellValue_1,
		},
		&pairs_types.Pairs{
			AccountType:  AccountType_2,
			Pair:         Pair_2,
			TargetSymbol: TargetSymbol_2,
			BaseSymbol:   BaseSymbol_2,
			Limit:        Limit_2,
			BuyDelta:     BuyDelta_2,
			BuyQuantity:  BuyQuantity_2,
			BuyValue:     BuyValue_2,
			SellDelta:    SellDelta_2,
			SellQuantity: SellQuantity_2,
			SellValue:    SellValue_2,
		},
	}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assert.NoError(t, err)
	assert.Equal(t, APIKey, config.GetAPIKey())
	assert.Equal(t, APISecret, config.GetSecretKey())
	assert.Equal(t, UseTestNet, config.GetUseTestNet())

	assert.Equal(t, (*checkingDate)[0].GetAccountType(), config.GetPair(Pair_1).GetAccountType())
	assert.Equal(t, (*checkingDate)[0].GetPair(), config.GetPair(Pair_1).GetPair())
	assert.Equal(t, (*checkingDate)[0].GetTargetSymbol(), config.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[0].GetBaseSymbol(), config.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[0].GetLimit(), config.GetPair(Pair_1).GetLimit())
	assert.Equal(t, (*checkingDate)[0].GetBuyDelta(), config.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[0].GetBuyQuantity(), config.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[0].GetBuyValue(), config.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, (*checkingDate)[0].GetSellDelta(), config.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, (*checkingDate)[0].GetSellQuantity(), config.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[0].GetSellValue(), config.GetPair(Pair_1).GetSellValue())

	assert.Equal(t, (*checkingDate)[1].GetAccountType(), config.GetPair(Pair_2).GetAccountType())
	assert.Equal(t, (*checkingDate)[1].GetPair(), config.GetPair(Pair_2).GetPair())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), config.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), config.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetLimit(), config.GetPair(Pair_2).GetLimit())
	assert.Equal(t, (*checkingDate)[1].GetBuyDelta(), config.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[1].GetBuyQuantity(), config.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[1].GetBuyValue(), config.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, (*checkingDate)[1].GetSellDelta(), config.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, (*checkingDate)[1].GetSellQuantity(), config.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[1].GetSellValue(), config.GetPair(Pair_2).GetSellValue())
}

func TestConfigGetPairs(t *testing.T) {
	config := &config_types.Configs{
		APIKey:     APIKey,
		APISecret:  APISecret,
		UseTestNet: UseTestNet,
		Pairs:      btree.New(2),
	}
	pairs := []config_interfaces.Pairs{
		&pairs_types.Pairs{
			AccountType:  AccountType_1,
			Pair:         Pair_1,
			TargetSymbol: TargetSymbol_1,
			BaseSymbol:   BaseSymbol_1,
			Limit:        Limit_1,
			BuyDelta:     BuyDelta_1,
			BuyQuantity:  BuyQuantity_1,
			BuyValue:     BuyValue_1,
			SellDelta:    SellDelta_1,
			SellQuantity: SellQuantity_1,
			SellValue:    SellValue_1,
		},
		&pairs_types.Pairs{
			AccountType:  AccountType_2,
			Pair:         Pair_2,
			TargetSymbol: TargetSymbol_2,
			BaseSymbol:   BaseSymbol_2,
			Limit:        Limit_2,
			BuyDelta:     BuyDelta_2,
			BuyQuantity:  BuyQuantity_2,
			BuyValue:     BuyValue_2,
			SellDelta:    SellDelta_2,
			SellQuantity: SellQuantity_2,
			SellValue:    SellValue_2,
		},
	}
	config.SetPairs(pairs)

	checkingDate, err := config.GetPairs()
	assert.NoError(t, err)
	assert.Equal(t, APIKey, config.GetAPIKey())
	assert.Equal(t, APISecret, config.GetSecretKey())
	assert.Equal(t, UseTestNet, config.GetUseTestNet())

	assert.Equal(t, (*checkingDate)[0].GetAccountType(), config.GetPair(Pair_1).GetAccountType())
	assert.Equal(t, (*checkingDate)[0].GetPair(), config.GetPair(Pair_1).GetPair())
	assert.Equal(t, (*checkingDate)[0].GetTargetSymbol(), config.GetPair(Pair_1).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[0].GetBaseSymbol(), config.GetPair(Pair_1).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[0].GetLimit(), config.GetPair(Pair_1).GetLimit())
	assert.Equal(t, (*checkingDate)[0].GetBuyDelta(), config.GetPair(Pair_1).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[0].GetBuyQuantity(), config.GetPair(Pair_1).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[0].GetBuyValue(), config.GetPair(Pair_1).GetBuyValue())
	assert.Equal(t, (*checkingDate)[0].GetSellDelta(), config.GetPair(Pair_1).GetSellDelta())
	assert.Equal(t, (*checkingDate)[0].GetSellQuantity(), config.GetPair(Pair_1).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[0].GetSellValue(), config.GetPair(Pair_1).GetSellValue())

	assert.Equal(t, (*checkingDate)[1].GetAccountType(), config.GetPair(Pair_2).GetAccountType())
	assert.Equal(t, (*checkingDate)[1].GetPair(), config.GetPair(Pair_2).GetPair())
	assert.Equal(t, (*checkingDate)[1].GetTargetSymbol(), config.GetPair(Pair_2).GetTargetSymbol())
	assert.Equal(t, (*checkingDate)[1].GetBaseSymbol(), config.GetPair(Pair_2).GetBaseSymbol())
	assert.Equal(t, (*checkingDate)[1].GetLimit(), config.GetPair(Pair_2).GetLimit())
	assert.Equal(t, (*checkingDate)[1].GetBuyDelta(), config.GetPair(Pair_2).GetBuyDelta())
	assert.Equal(t, (*checkingDate)[1].GetBuyQuantity(), config.GetPair(Pair_2).GetBuyQuantity())
	assert.Equal(t, (*checkingDate)[1].GetBuyValue(), config.GetPair(Pair_2).GetBuyValue())
	assert.Equal(t, (*checkingDate)[1].GetSellDelta(), config.GetPair(Pair_2).GetSellDelta())
	assert.Equal(t, (*checkingDate)[1].GetSellQuantity(), config.GetPair(Pair_2).GetSellQuantity())
	assert.Equal(t, (*checkingDate)[1].GetSellValue(), config.GetPair(Pair_2).GetSellValue())
}
