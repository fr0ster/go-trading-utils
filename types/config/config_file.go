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
		cf.Configs.Pairs.ReplaceOrInsert(&pairs_types.Pairs{
			InitialBalance:         0.0,
			AccountType:            "SPOT/MARGIN/ISOLATED_MARGIN/USDT_FUTURE/COIN_FUTURE",
			StrategyType:           "HOLDING/SCALPING/ARBITRAGE/TRADING",
			StageType:              "INPUT_INTO_POSITION/WORK_IN_POSITION/OUTPUT_OF_POSITION",
			Pair:                   "BTCUSDT",
			TargetSymbol:           "BTC",
			BaseSymbol:             "USDT",
			LimitInputIntoPosition: 0.1,
			LimitOnPosition:        1.0,
			LimitOnTransaction:     0.01,
			BuyDelta:               0.01,
			BuyQuantity:            0.0,
			BuyValue:               0.0,
			SellDelta:              0.05,
			SellQuantity:           0.0,
			SellValue:              0.0,
			Commission:             []pairs_types.Commission{},
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
	res = &ConfigFile{
		FilePath: file_path,
		Configs: &Configs{
			APIKey:     "",
			APISecret:  "",
			UseTestNet: false,
			Pairs:      btree.New(degree),
		},
	}
	return
}
