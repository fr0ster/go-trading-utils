package utils_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
)

func TestConfig_Load(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Create a sample config
	time, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	sampleConfig := types.Config{
		Timestamp:         time,
		AccountType:       "SPOT",
		Symbol:            "BTCUSDT",
		Balance:           1000.0,
		CalculatedBalance: 1000.0,
		Quantity:          1.0,
		Value:             1000.0,
		BoundQuantity:     1.0,
	}

	// Write the sample config to the temporary file
	configBytes, err := json.Marshal(sampleConfig)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(tempFile.Name(), configBytes, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new Config instance
	config := utils.NewConfig(tempFile.Name())

	// Load the config from the file
	err = config.Load()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the loaded config matches the sample config
	if config.Config != sampleConfig {
		t.Errorf("Expected config to be %v, but got %v", sampleConfig, config.Config)
	}
}

func TestConfig_Save(t *testing.T) {
	// Create a temporary directory for testing
	tmpfile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpfile.Name())

	// Create a sample config
	time, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	sampleConfig := types.Config{
		Timestamp:         time,
		AccountType:       "SPOT",
		Symbol:            "BTCUSDT",
		Balance:           1000.0,
		CalculatedBalance: 1000.0,
		Quantity:          1.0,
		Value:             1000.0,
		BoundQuantity:     1.0,
	}

	// Create a new Config instance
	config := utils.NewConfig(tmpfile.Name())
	config.Config = sampleConfig

	// Save the config to the file
	err = config.Save()
	if err != nil {
		t.Fatal(err)
	}

	// Read the saved config from the file
	savedConfigBytes, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal the saved config
	var savedConfig types.Config
	err = json.Unmarshal(savedConfigBytes, &savedConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the saved config matches the sample config
	if savedConfig != sampleConfig {
		t.Errorf("Expected saved config to be %v, but got %v", sampleConfig, savedConfig)
	}
}
