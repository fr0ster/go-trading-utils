package config

import (
	"io"
	"os"
	"sync"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
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
	if cf.Configs.Pairs == nil {
		cf.Configs.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
			Pair:         "BTCUSDT",
			TargetSymbol: "BTC",
			BaseSymbol:   "USDT",
			Limit:        10.0,
			Quantity:     1.0,
			Value:        100.0,
		})
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
	pairBtcUsdt := &pairs_types.Pairs{
		AccountType:  pairs_types.SpotAccountType,
		Pair:         "BTCUSDT",
		TargetSymbol: "BTC",
		BaseSymbol:   "USDT",
		Limit:        0.01,
		Delta:        0.01,
		Quantity:     0.0,
		Value:        0.0,
	}
	res = &ConfigFile{
		FilePath: file_path,
		Configs: &Configs{
			APIKey:     "",
			APISecret:  "",
			UseTestNet: false,
			Pairs:      btree.New(degree),
		},
	}
	res.Configs.Pairs.ReplaceOrInsert(pairBtcUsdt)
	return
}
