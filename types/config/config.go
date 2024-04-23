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
		SpotConnection    *connection_types.Connection `json:"spot_connection"`
		FuturesConnection *connection_types.Connection `json:"futures_connection"`
		LogLevel          logrus.Level                 `json:"log_level"`
		Pairs             *btree.BTree
	}
)

// GetFuturesConnection implements config.Configuration.
func (cf *Configs) GetFuturesConnection() connection_interfaces.Connection {
	return cf.FuturesConnection
}

// GetSpotConnection implements config.Configuration.
func (cf *Configs) GetSpotConnection() connection_interfaces.Connection {
	return cf.SpotConnection
}

func (cf *Configs) GetLogLevel() logrus.Level {
	return cf.LogLevel
}

func (cf *Configs) SetLogLevel(level logrus.Level) {
	cf.LogLevel = level
}

// Implement the GetPair method
func (cf *Configs) GetPair(pair string) pairs_interfaces.Pairs {
	res := cf.Pairs.Get(&pairs_types.Pairs{Pair: pair})
	return res.(*pairs_types.Pairs)
}

// Implement the GetPairs method
func (cf *Configs) GetPairs(account_type ...pairs_types.AccountType) (*[]pairs_interfaces.Pairs, error) {
	isExist := func(a pairs_types.AccountType) bool {
		for _, at := range account_type {
			if at == a {
				return true
			}
		}
		return false
	}
	pairs := make([]pairs_interfaces.Pairs, 0)
	cf.Pairs.Ascend(func(a btree.Item) bool {
		if len(account_type) == 0 || isExist(a.(*pairs_types.Pairs).AccountType) {
			pairs = append(pairs, a.(*pairs_types.Pairs))
		}
		return true
	})
	if len(pairs) == 0 {
		return nil, errors.New("no pairs found in the configuration file")
	}
	return &pairs, nil
}

// Implement the SetPairs method
func (cf *Configs) SetPairs(pairs []pairs_interfaces.Pairs) error {
	for _, pair := range pairs {
		cf.Pairs.ReplaceOrInsert(pair.(*pairs_types.Pairs))
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
		SpotConnection    *connection_types.Connection `json:"spot_connection"`
		FuturesConnection *connection_types.Connection `json:"futures_connection"`
		LogLevel          string                       `json:"log_level"`
		Pairs             []*pairs_types.Pairs         `json:"pairs"`
	}{
		SpotConnection:    c.SpotConnection,
		FuturesConnection: c.FuturesConnection,
		LogLevel:          c.LogLevel.String(),
		Pairs:             pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		SpotConnection    *connection_types.Connection `json:"spot_connection"`
		FuturesConnection *connection_types.Connection `json:"futures_connection"`
		LogLevel          string                       `json:"log_level"`
		Pairs             []*pairs_types.Pairs         `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.SpotConnection = &connection_types.Connection{
		APIKey:          temp.SpotConnection.APIKey,
		APISecret:       temp.SpotConnection.APISecret,
		UseTestNet:      temp.SpotConnection.UseTestNet,
		CommissionMaker: temp.SpotConnection.CommissionMaker,
		CommissionTaker: temp.SpotConnection.CommissionTaker,
	}
	c.FuturesConnection = &connection_types.Connection{
		APIKey:          temp.FuturesConnection.APIKey,
		APISecret:       temp.FuturesConnection.APISecret,
		UseTestNet:      temp.FuturesConnection.UseTestNet,
		CommissionMaker: temp.FuturesConnection.CommissionMaker,
		CommissionTaker: temp.FuturesConnection.CommissionTaker,
	}
	// Parse the string log level to a logrus.Level
	var err error
	c.LogLevel, err = logrus.ParseLevel(temp.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", temp.LogLevel)
	}
	c.Pairs = btree.New(2)
	for _, pair := range temp.Pairs {
		if pair.Connection == nil {
			pair.Connection = &connection_types.Connection{}
		}
		if pair.AccountType == pairs_types.SpotAccountType &&
			pair.Connection.GetAPIKey() == "" &&
			pair.Connection.GetSecretKey() == "" &&
			!pair.Connection.GetUseTestNet() &&
			pair.Connection.GetCommissionMaker() == 0 &&
			pair.Connection.GetCommissionTaker() == 0 {
			pair.Connection = c.SpotConnection
		} else if pair.AccountType == pairs_types.USDTFutureType &&
			pair.Connection.GetAPIKey() == "" &&
			pair.Connection.GetSecretKey() == "" &&
			!pair.Connection.GetUseTestNet() &&
			pair.Connection.GetCommissionMaker() == 0 &&
			pair.Connection.GetCommissionTaker() == 0 {
			pair.Connection = c.FuturesConnection
		}

		c.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}
