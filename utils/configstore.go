package utils

import (
	"encoding/json"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/types"
)

type Config struct {
	Config   types.Config
	FilePath string
}

func NewConfig(filePath string) *Config {
	config := &Config{
		Config:   types.Config{},
		FilePath: filePath,
	}
	err := config.Load()
	if err != nil {
		config = &Config{
			Config: types.Config{
				AccountType:   binance.AccountTypeSpot,
				Symbol:        "BTCUSDT",
				Balance:       0.0,
				Value:         0.0,
				Quantity:      0.0,
				BoundQuantity: 0.0,
			}, FilePath: filePath}
	}
	return config
}

func (c *Config) Load() error {
	ds := NewDataStore(c.FilePath)
	err := ds.LoadFromFile()
	if err != nil {
		return err
	}
	err = json.Unmarshal(ds.GetData(), &c.Config)
	return err
}

func (c *Config) Save() (err error) {
	ds := NewDataStore(c.FilePath)
	jsonBytes, err := json.Marshal(c.Config)
	if err != nil {
		return err
	}
	ds.SetData(jsonBytes)
	ds.SaveToFile()
	return nil
}
