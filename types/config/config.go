package config

import (
	"encoding/json"
	"errors"

	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	pairs_types "github.com/fr0ster/go-trading-utils/types/config/pairs"
	"github.com/google/btree"
)

type (
	Configs struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
		Pairs      *btree.BTree
	}
)

func (cf *Configs) GetAPIKey() string {
	return cf.APIKey
}

func (cf *Configs) GetSecretKey() string {
	return cf.APISecret
}

func (cf *Configs) GetUseTestNet() bool {
	return cf.UseTestNet
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
		APIKey     string               `json:"api_key"`
		APISecret  string               `json:"api_secret"`
		UseTestNet bool                 `json:"use_test_net"`
		Pairs      []*pairs_types.Pairs `json:"pairs"`
	}{
		APIKey:     c.APIKey,
		APISecret:  c.APISecret,
		UseTestNet: c.UseTestNet,
		Pairs:      pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		APIKey     string               `json:"api_key"`
		APISecret  string               `json:"api_secret"`
		UseTestNet bool                 `json:"use_test_net"`
		Pairs      []*pairs_types.Pairs `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.APIKey = temp.APIKey
	c.APISecret = temp.APISecret
	c.UseTestNet = temp.UseTestNet
	c.Pairs = btree.New(2)
	for _, pair := range temp.Pairs {
		c.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}
