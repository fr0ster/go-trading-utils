package config

import (
	"io"
	"os"
	"sync"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
)

type (
	ConfigFile struct {
		FilePath string   `json:"file_path"`
		Configs  *Configs `json:"symbols"`
		mu       sync.Mutex
	}
)

func (cf *ConfigFile) Lock() {
	cf.mu.Lock()
}

func (cf *ConfigFile) Unlock() {
	cf.mu.Unlock()
}

func (cf *ConfigFile) Load() error {
	cf.Lock()
	defer cf.Unlock()
	file, err := os.Open(cf.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = cf.Configs.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ConfigFile) Save() error {
	cf.Lock()
	defer cf.Unlock()
	if cf.Configs.Pairs.Len() == 0 {
		cf.Configs.Pairs.ReplaceOrInsert(
			pairs_types.New(
				&connection_types.Connection{},
				pairs_types.SpotAccountType,
				pairs_types.HoldingStrategyType,
				pairs_types.InputIntoPositionStage,
				"BTCUSDT",
				"BTC",
				"USDT"))
	}

	formattedJSON, err := cf.Configs.MarshalJSON()
	if err != nil {
		return err
	}

	err = os.WriteFile(cf.FilePath, formattedJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ConfigFile) GetConfigurations() config_types.Configuration {
	return cf.Configs
}

// New creates a new ConfigRecord with the provided API key, API secret, and symbols.
func ConfigNew(file_path string, degree int) (res *ConfigFile) {
	res = &ConfigFile{
		FilePath: file_path,
		Configs: &Configs{
			SpotConnection: &connection_types.Connection{
				APIKey:          "",
				APISecret:       "",
				UseTestNet:      false,
				CommissionMaker: 0.001,
				CommissionTaker: 0.001,
			},
			FuturesConnection: &connection_types.Connection{
				APIKey:          "",
				APISecret:       "",
				UseTestNet:      false,
				CommissionMaker: 0.001,
				CommissionTaker: 0.001,
			},
			LogLevel: 0x00,
			Pairs:    btree.New(degree),
		},
	}
	return
}
