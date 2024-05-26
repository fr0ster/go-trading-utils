package config

import (
	"io"
	"os"
	"sync"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
)

type (
	ConfigFile struct {
		filePath string
		configs  *Configs
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
	file, err := os.Open(cf.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = cf.configs.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ConfigFile) Save() error {
	cf.Lock()
	defer cf.Unlock()
	if cf.configs.Pairs.Len() == 0 {
		cf.configs.Pairs.ReplaceOrInsert(
			pairs_types.New(
				&connection_types.Connection{},
				pairs_types.SpotAccountType,
				pairs_types.HoldingStrategyType,
				pairs_types.InputIntoPositionStage,
				"BTCUSDT",
				"BTC",
				"USDT"))
	}

	formattedJSON, err := cf.configs.MarshalJSON()
	if err != nil {
		return err
	}

	err = os.WriteFile(cf.filePath, formattedJSON, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ConfigFile) GetConfigurations() config_types.Configuration {
	return cf.configs
}

func (cf *ConfigFile) SetConfigurations(config config_types.Configuration) {
	cf.configs = config.(*Configs)
}

// New creates a new ConfigRecord with the provided API key, API secret, and symbols.
func NewConfigFile(
	file_path string,
	degree int) (res *ConfigFile) {
	res = &ConfigFile{
		filePath: file_path,
		configs:  NewConfig(&connection_types.Connection{}),
	}
	return
}
