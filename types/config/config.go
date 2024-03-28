package config

import (
	"encoding/json"
	"os"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
)

type (
	Configs struct {
		APIKey       string  `json:"api_key"`
		APISecret    string  `json:"api_secret"`
		UseTestNet   bool    `json:"use_test_net"`
		Pair         string  `json:"symbol"`
		TargetSymbol string  `json:"target_symbol"`
		BaseSymbol   string  `json:"base_symbol"`
		Limit        float64 `json:"limit"`
		Quantity     float64 `json:"quantity"`
		Value        float64 `json:"value"`
	}
	ConfigFile struct {
		FilePath string   `json:"file_path"`
		Configs  *Configs `json:"configs"`
	}
)

// New creates a new ConfigRecord with the provided API key, API secret, and symbols.
func ConfigNew(file_path string) *ConfigFile {
	return &ConfigFile{
		FilePath: file_path,
		Configs: &Configs{
			APIKey:       "",
			APISecret:    "",
			Pair:         "BTCUSDT",
			TargetSymbol: "BTC",
			BaseSymbol:   "USDT",
			Quantity:     0,
			Value:        0,
		},
	}
}

func (cr *ConfigFile) Load() error {
	// Check if file exists
	if _, err := os.Stat(cr.FilePath); os.IsNotExist(err) {
		// File does not exist, create a new one with default config
		return err
	}

	// Open the file
	file, err := os.Open(cr.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the JSON config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cr.Configs)
	if err != nil {
		return err
	}
	return nil
}

func (cr *ConfigFile) Save() error {
	formattedJSON, err := json.MarshalIndent(cr.Configs, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cr.FilePath, formattedJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

// GetSymbol implements Configuration.
func (cr *Configs) GetPair() string {
	return cr.Pair
}

// GetBaseSymbol implements config.Configuration.
func (cr *Configs) GetBaseSymbol() string {
	return cr.BaseSymbol
}

// GetTargetSymbol implements config.Configuration.
func (cr *Configs) GetTargetSymbol() string {
	return cr.TargetSymbol
}

func (cr *Configs) GetLimit() float64 {
	return cr.Limit
}

func (cr *Configs) GetQuantity() float64 {
	return cr.Quantity
}

func (cr *Configs) GetValue() float64 {
	return cr.Value
}

func (cr *Configs) GetAPIKey() string {
	return cr.APIKey
}

func (cr *Configs) GetSecretKey() string {
	return cr.APISecret
}

func (cr *Configs) GetUseTestNet() bool {
	return cr.UseTestNet
}

func (cr *ConfigFile) GetConfigurations() config_types.Configuration {
	return cr.Configs
}
