package config

import (
	"encoding/json"
	"errors"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	"github.com/google/btree"
)

type (
	Connection struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}
	Configs struct {
		SpotConnection    *Connection `json:"spot_connection"`
		FuturesConnection *Connection `json:"futures_connection"`
		Pairs             *btree.BTree
	}
)

func (cf *Connection) GetAPIKey() string {
	return cf.APIKey
}

func (cf *Connection) SetApiKey(key string) {
	cf.APIKey = key
}

func (cf *Connection) GetSecretKey() string {
	return cf.APISecret
}

func (cf *Connection) SetSecretKey(key string) {
	cf.APISecret = key
}

func (cf *Connection) GetUseTestNet() bool {
	return cf.UseTestNet
}

func (cf *Connection) SetUseTestNet(useTestNet bool) {
	cf.UseTestNet = useTestNet
}

func (cf *Connection) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}{
		APIKey:     cf.APIKey,
		APISecret:  cf.APISecret,
		UseTestNet: cf.UseTestNet,
	})
}

func (cf *Connection) UnmarshalJSON(data []byte) error {
	temp := &struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	cf.APIKey = temp.APIKey
	cf.APISecret = temp.APISecret
	cf.UseTestNet = temp.UseTestNet
	return nil
}

// GetFuturesConnection implements config.Configuration.
func (cf *Configs) GetFuturesConnection() config_interfaces.Connection {
	return cf.FuturesConnection
}

// GetSpotConnection implements config.Configuration.
func (cf *Configs) GetSpotConnection() config_interfaces.Connection {
	return cf.SpotConnection
}

func (cf *Configs) GetPair(pair string) config_interfaces.Pairs {
	// Implement the GetPair method
	res := cf.Pairs.Get(&pairs_types.Pairs{Pair: pair})
	return res.(*pairs_types.Pairs)
}

func (cf *Configs) GetPairs(account_type ...pairs_types.AccountType) (*[]config_interfaces.Pairs, error) {
	// Implement the GetPairs method
	isExist := func(a pairs_types.AccountType) bool {
		for _, at := range account_type {
			if at == a {
				return true
			}
		}
		return false
	}
	pairs := make([]config_interfaces.Pairs, 0)
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

func (cf *Configs) SetPairs(pairs []config_interfaces.Pairs) error {
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
		SpotConnection    *Connection          `json:"spot_connection"`
		FuturesConnection *Connection          `json:"futures_connection"`
		Pairs             []*pairs_types.Pairs `json:"pairs"`
	}{
		SpotConnection:    c.SpotConnection,
		FuturesConnection: c.FuturesConnection,
		Pairs:             pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		SpotConnection    *Connection          `json:"spot_connection"`
		FuturesConnection *Connection          `json:"futures_connection"`
		Pairs             []*pairs_types.Pairs `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.SpotConnection = &Connection{
		APIKey:     temp.SpotConnection.APIKey,
		APISecret:  temp.SpotConnection.APISecret,
		UseTestNet: temp.SpotConnection.UseTestNet,
	}
	c.FuturesConnection = &Connection{
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
