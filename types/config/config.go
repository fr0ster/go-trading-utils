package config

import (
	"encoding/json"
	"errors"
	"fmt"

	connection_interfaces "github.com/fr0ster/go-trading-utils/interfaces/connection"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"
	"github.com/sirupsen/logrus"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
)

type (
	Configs struct {
		Connection   *connection_types.Connection `json:"connection"`
		LogLevel     logrus.Level                 `json:"log_level"`
		ReloadConfig bool                         `json:"reload_config"`
		Pairs        *btree.BTree
	}
)

// GetSpotConnection implements config.Configuration.
func (cf *Configs) GetConnection() connection_interfaces.Connection {
	return cf.Connection
}

func (cf *Configs) GetLogLevel() logrus.Level {
	return cf.LogLevel
}

func (cf *Configs) SetLogLevel(level logrus.Level) {
	cf.LogLevel = level
}

func (cf *Configs) GetReloadConfig() bool {
	return cf.ReloadConfig
}

// Implement the GetPair method
func (cf *Configs) GetPair(pair string) pairs_interfaces.Pairs {
	res := cf.Pairs.Get(&pairs_types.Pairs{Pair: pair})
	return res.(*pairs_types.Pairs)
}

// Implement the SetPair method
func (cf *Configs) SetPair(pair pairs_interfaces.Pairs) {
	cf.Pairs.ReplaceOrInsert(pair.(*pairs_types.Pairs))
}

// Implement the GetPairs method
func (cf *Configs) GetPairs(account_type ...pairs_types.AccountType) ([]*pairs_types.Pairs, error) {
	isExist := func(a pairs_types.AccountType) bool {
		for _, at := range account_type {
			if at == a {
				return true
			}
		}
		return false
	}
	pairs := make([]*pairs_types.Pairs, 0)
	cf.Pairs.Ascend(func(a btree.Item) bool {
		if len(account_type) == 0 || isExist(a.(*pairs_types.Pairs).AccountType) {
			pairs = append(pairs, a.(*pairs_types.Pairs))
		}
		return true
	})
	if len(pairs) == 0 {
		return nil, errors.New("no pairs found in the configuration file")
	}
	return pairs, nil
}

// Implement the SetPairs method
func (cf *Configs) SetPairs(pairs []*pairs_types.Pairs) error {
	for _, pair := range pairs {
		cf.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}

func (c *Configs) MarshalJSON() ([]byte, error) {
	pairs := make([]*pairs_types.Pairs, 0)
	c.Pairs.Ascend(func(a btree.Item) bool {
		pairs = append(pairs, a.(*pairs_types.Pairs))
		return true
	})
	return json.MarshalIndent(&struct {
		Connection   *connection_types.Connection `json:"connection"`
		LogLevel     string                       `json:"log_level"`
		ReloadConfig bool                         `json:"reload_config"`
		Pairs        []*pairs_types.Pairs         `json:"pairs"`
	}{
		Connection:   c.Connection,
		LogLevel:     c.LogLevel.String(),
		ReloadConfig: c.ReloadConfig,
		Pairs:        pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		Connection   *connection_types.Connection `json:"connection"`
		LogLevel     string                       `json:"log_level"`
		ReloadConfig bool                         `json:"reload_config"`
		Pairs        []*pairs_types.Pairs         `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.Connection = &connection_types.Connection{
		APIKey:          temp.Connection.APIKey,
		APISecret:       temp.Connection.APISecret,
		UseTestNet:      temp.Connection.UseTestNet,
		CommissionMaker: temp.Connection.CommissionMaker,
		CommissionTaker: temp.Connection.CommissionTaker,
	}
	// Parse the string log level to a logrus.Level
	var err error
	c.LogLevel, err = logrus.ParseLevel(temp.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", temp.LogLevel)
	}
	c.ReloadConfig = temp.ReloadConfig
	c.Pairs = btree.New(2)
	for _, pair := range temp.Pairs {
		c.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}

func NewConfig(connection *connection_types.Connection) *Configs {
	return &Configs{
		Connection:   connection,
		LogLevel:     logrus.InfoLevel,
		ReloadConfig: false,
		Pairs:        btree.New(2),
	}
}
