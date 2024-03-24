package config_test

import (
	"encoding/json"
	"os"
	"testing"

	config_types "github.com/fr0ster/go-trading-utils/types/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigFile_Load(t *testing.T) {
	// Create a temporary config file for testing
	tmpFile, err := os.CreateTemp("", "config.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write test data to the temporary config file
	testData := []byte(`{
			"api_key": "your_api_key",
			"api_secret": "your_api_secret",
			"symbol": "BTCUSDT",
			"limit": 10.0,
			"quantity": 1.0,
			"value": 100.0
		}`)
	err = os.WriteFile(tmpFile.Name(), testData, 0644)
	assert.NoError(t, err)

	// Create a new ConfigFile instance
	config := config_types.ConfigNew(tmpFile.Name())

	// Load the config from the file
	err = config.Load()
	assert.NoError(t, err)

	// Assert that the loaded config matches the test data
	assert.Equal(t, "your_api_key", config.Configs.APIKey)
	assert.Equal(t, "your_api_secret", config.Configs.APISecret)
	assert.Equal(t, "BTCUSDT", config.Configs.Symbol)
	assert.Equal(t, 10.0, config.Configs.Limit)
	assert.Equal(t, 1.0, config.Configs.Quantity)
	assert.Equal(t, 100.0, config.Configs.Value)
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
			APIKey:    "your_api_key",
			APISecret: "your_api_secret",
			Symbol:    "BTCUSDT",
			Limit:     10.0,
			Quantity:  1.0,
			Value:     100.0,
		},
	}

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
	assert.Equal(t, config.Configs.APIKey, savedConfig.APIKey)
	assert.Equal(t, config.Configs.APISecret, savedConfig.APISecret)
	assert.Equal(t, config.Configs.Symbol, savedConfig.Symbol)
	assert.Equal(t, config.Configs.Limit, savedConfig.Limit)
	assert.Equal(t, config.Configs.Quantity, savedConfig.Quantity)
	assert.Equal(t, config.Configs.Value, savedConfig.Value)
}

// Add more tests for other methods if needed
