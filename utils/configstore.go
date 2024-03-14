package utils

import (
	"encoding/json"

	"github.com/fr0ster/go-binance-utils/types"
)

type Config struct {
	Config   types.Config
	FilePath string
}

func NewConfig(filePath string) *Config {
	return &Config{
		Config:   types.Config{},
		FilePath: filePath,
	}
}

func (c *Config) Load() error {
	ds := NewDataStore(c.FilePath)
	err := ds.LoadFromFile()
	if err != nil {
		c.Config = types.Config{}
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
