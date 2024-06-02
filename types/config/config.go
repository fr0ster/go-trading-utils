package config

import (
	"encoding/json"
	"errors"
	"fmt"

	connection_interfaces "github.com/fr0ster/go-trading-utils/interfaces/connection"
	"github.com/sirupsen/logrus"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
)

type (
	Configs struct {
		Connection                    *connection_types.Connection `json:"connection"`
		LogLevel                      logrus.Level                 `json:"log_level"`
		ReloadConfig                  bool                         `json:"reload_config"`
		ObservePriceLiquidation       bool                         `json:"observe_price_liquidation"`
		PercentsToLiquidation         float64                      `json:"percents_to_liquidation"`
		PercentToDecreasePosition     float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOut               int                          `json:"observer_timeout"`
		MaintainPartiallyFilledOrders bool                         `json:"maintain_partially_filled_orders"`
		Pairs                         *btree.BTree
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

func (cf *Configs) GetObservePriceLiquidation() bool {
	return cf.ObservePriceLiquidation
}

func (cf *Configs) GetPercentsToLiquidation() float64 {
	return cf.PercentsToLiquidation
}

func (cf *Configs) GetPercentToDecreasePosition() float64 {
	return cf.PercentToDecreasePosition
}

func (cf *Configs) GetObserverTimeOut() int {
	return cf.ObserverTimeOut
}

func (cf *Configs) GetMaintainPartiallyFilledOrders() bool {
	return cf.MaintainPartiallyFilledOrders
}

// Implement the GetPair method
func (cf *Configs) GetPair(
	account pairs_types.AccountType,
	strategy pairs_types.StrategyType,
	stage pairs_types.StageType,
	pair string) *pairs_types.Pairs {
	if res := cf.Pairs.Get(&pairs_types.Pairs{
		AccountType:  account,
		StrategyType: strategy,
		StageType:    stage,
		Pair:         pair}); res != nil {
		return res.(*pairs_types.Pairs)
	} else {
		return nil
	}
}

// Implement the SetPair method
func (cf *Configs) SetPair(pair *pairs_types.Pairs) {
	cf.Pairs.ReplaceOrInsert(pair)
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
		Connection                    *connection_types.Connection `json:"connection"`
		LogLevel                      string                       `json:"log_level"`
		ReloadConfig                  bool                         `json:"reload_config"`
		ObservePriceLiquidation       bool                         `json:"observe_price_liquidation"`
		PercentsToLiquidation         float64                      `json:"percents_to_liquidation"`
		PercentToDecreasePosition     float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOut               int                          `json:"observer_timeout"`
		MaintainPartiallyFilledOrders bool                         `json:"maintain_partially_filled_orders"`
		Pairs                         []*pairs_types.Pairs         `json:"pairs"`
	}{
		Connection:                    c.Connection,
		LogLevel:                      c.LogLevel.String(),
		ReloadConfig:                  c.ReloadConfig,
		ObservePriceLiquidation:       c.ObservePriceLiquidation,
		PercentsToLiquidation:         c.PercentsToLiquidation,
		PercentToDecreasePosition:     c.PercentToDecreasePosition,
		ObserverTimeOut:               c.ObserverTimeOut,
		MaintainPartiallyFilledOrders: c.MaintainPartiallyFilledOrders,
		Pairs:                         pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		Connection                    *connection_types.Connection `json:"connection"`
		LogLevel                      string                       `json:"log_level"`
		ReloadConfig                  bool                         `json:"reload_config"`
		ObservePriceLiquidation       bool                         `json:"observe_price_liquidation"`
		PercentsToLiquidation         float64                      `json:"percents_to_liquidation"`
		PercentToDecreasePosition     float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOut               int                          `json:"observer_timeout"`
		MaintainPartiallyFilledOrders bool                         `json:"maintain_partially_filled_orders"`
		Pairs                         []*pairs_types.Pairs         `json:"pairs"`
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
	c.ObservePriceLiquidation = temp.ObservePriceLiquidation
	c.PercentsToLiquidation = temp.PercentsToLiquidation
	c.PercentToDecreasePosition = temp.PercentToDecreasePosition
	c.ObserverTimeOut = temp.ObserverTimeOut
	c.MaintainPartiallyFilledOrders = temp.MaintainPartiallyFilledOrders
	if c.Pairs == nil || c.Pairs.Len() == 0 {
		c.Pairs = btree.New(2)
	}
	for _, pair := range temp.Pairs {
		c.Pairs.ReplaceOrInsert(pair)
	}
	return nil
}

func NewConfig(connection *connection_types.Connection) *Configs {
	return &Configs{
		Connection:                    connection,
		LogLevel:                      logrus.InfoLevel,
		ReloadConfig:                  false,
		ObservePriceLiquidation:       false,
		PercentsToLiquidation:         0.05,
		PercentToDecreasePosition:     0.03,
		ObserverTimeOut:               1000,
		MaintainPartiallyFilledOrders: false,
		Pairs:                         btree.New(2),
	}
}
