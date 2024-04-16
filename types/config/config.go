package config

import (
	"encoding/json"
	"errors"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	pairs_interfaces "github.com/fr0ster/go-trading-utils/interfaces/pairs"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
)

type (
	Configs struct {
		SpotConnection    *connection_types.Connection `json:"spot_connection"`
		FuturesConnection *connection_types.Connection `json:"futures_connection"`
		Pairs             *btree.BTree
	}
)

// GetFuturesConnection implements config.Configuration.
func (cf *Configs) GetFuturesConnection() config_interfaces.Connection {
	return cf.FuturesConnection
}

// GetSpotConnection implements config.Configuration.
func (cf *Configs) GetSpotConnection() config_interfaces.Connection {
	return cf.SpotConnection
}

func (cf *Configs) GetPair(pair string) pairs_interfaces.Pairs {
	// Implement the GetPair method
	res := cf.Pairs.Get(&pairs_types.Pairs{Pair: pair})
	return res.(*pairs_types.Pairs)
}

func (cf *Configs) GetPairs(account_type ...pairs_types.AccountType) (*[]pairs_interfaces.Pairs, error) {
	// Implement the GetPairs method
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

func (cf *Configs) SetPairs(pairs []pairs_interfaces.Pairs) error {
	// Implement the SetPairs method
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
		Pairs             []*pairs_types.Pairs         `json:"pairs"`
	}{
		SpotConnection:    c.SpotConnection,
		FuturesConnection: c.FuturesConnection,
		Pairs:             pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		SpotConnection    *connection_types.Connection `json:"spot_connection"`
		FuturesConnection *connection_types.Connection `json:"futures_connection"`
		Pairs             []*pairs_types.Pairs         `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.SpotConnection = &connection_types.Connection{
		APIKey:     temp.SpotConnection.APIKey,
		APISecret:  temp.SpotConnection.APISecret,
		UseTestNet: temp.SpotConnection.UseTestNet,
	}
	c.FuturesConnection = &connection_types.Connection{
		APIKey:     temp.FuturesConnection.APIKey,
		APISecret:  temp.FuturesConnection.APISecret,
		UseTestNet: temp.FuturesConnection.UseTestNet,
	}
	c.Pairs = btree.New(2)
	for _, pair := range temp.Pairs {
		c.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}
