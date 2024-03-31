package config

import (
	"io"
	"os"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
	"github.com/google/btree"
)

type (
	ConfigFile struct {
		FilePath string  `json:"file_path"`
		Configs  Configs `json:"symbols"`
	}
)

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
func ConfigNew(file_path string, degree int) *ConfigFile {
	return &ConfigFile{
		FilePath: file_path,
		Configs: Configs{
			APIKey:     "",
			APISecret:  "",
			UseTestNet: false,
			Pairs:      btree.New(degree)},
	}
}
