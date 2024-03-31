package config

import (
	"encoding/json"

	config_types "github.com/fr0ster/go-trading-utils/interfaces/config"
	"github.com/google/btree"
)

type (
	PairsSlice []*Pairs
	Configs    struct {
		APIKey     string `json:"api_key"`
		APISecret  string `json:"api_secret"`
		UseTestNet bool   `json:"use_test_net"`
		Pairs      *btree.BTree
	}
)

func (p *PairsSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal([]*Pairs(*p))
}

func (p *PairsSlice) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, (*[]*Pairs)(p))
}

func (cf Configs) GetAPIKey() string {
	return cf.APIKey
}

func (cf Configs) GetSecretKey() string {
	return cf.APISecret
}

func (cf Configs) GetUseTestNet() bool {
	return cf.UseTestNet
}

func (cf Configs) GetPairs(pair string) config_types.Pairs {
	// Implement the GetPair method
	res := cf.Pairs.Get(&Pairs{Pair: pair})
	return res.(*Pairs)
}

func (c *Configs) MarshalJSON() ([]byte, error) {
	pairs := make(PairsSlice, 0)
	c.Pairs.Ascend(func(a btree.Item) bool {
		pairs = append(pairs, a.(*Pairs))
		return true
	})
	return json.Marshal(&struct {
		APIKey     string     `json:"api_key"`
		APISecret  string     `json:"api_secret"`
		UseTestNet bool       `json:"use_test_net"`
		Pairs      PairsSlice `json:"pairs"`
	}{
		APIKey:     c.APIKey,
		APISecret:  c.APISecret,
		UseTestNet: c.UseTestNet,
		Pairs:      pairs,
	})
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		APIKey     string     `json:"api_key"`
		APISecret  string     `json:"api_secret"`
		UseTestNet bool       `json:"use_test_net"`
		Pairs      PairsSlice `json:"pairs"`
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
