package config

import (
	"encoding/json"
	"errors"
	"fmt"

	connection_interfaces "github.com/fr0ster/go-trading-utils/interfaces/connection"
	"github.com/sirupsen/logrus"

	connection_types "github.com/fr0ster/go-trading-utils/types/connection"
	items_types "github.com/fr0ster/go-trading-utils/types/depth/items"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"

	"github.com/google/btree"
)

type (
	Configs struct {
		Connection                    *connection_types.Connection `json:"connection"`
		LogLevel                      logrus.Level                 `json:"log_level"`
		ObservePriceLiquidation       bool                         `json:"observe_price_liquidation"`
		ObservePosition               bool                         `json:"observe_position"`
		ClosePositionOnRestart        bool                         `json:"close_position_on_restart"`
		BalancingOfMargin             bool                         `json:"balancing_of_margin"`
		PercentsToStopSettingNewOrder items_types.PricePercentType `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition     items_types.PricePercentType `json:"percent_to_decrease_position"`
		ObserverTimeOutMillisecond    int                          `json:"observer_timeout_millisecond"`
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

func (cf *Configs) GetObservePriceLiquidation() bool {
	return cf.ObservePriceLiquidation
}

func (cf *Configs) SetObservePriceLiquidation(observe bool) {
	cf.ObservePriceLiquidation = observe
}

func (cf *Configs) GetObservePosition() bool {
	return cf.ObservePosition
}

func (cf *Configs) SetObservePosition(observe bool) {
	cf.ObservePosition = observe
}

func (cf *Configs) GetClosePositionOnRestart() bool {
	return cf.ClosePositionOnRestart
}

func (cf *Configs) SetClosePositionOnRestart(close bool) {
	cf.ClosePositionOnRestart = close
}

func (cf *Configs) GetBalancingOfMargin() bool {
	return cf.BalancingOfMargin
}

func (cf *Configs) SetBalancingOfMargin(balancing bool) {
	cf.BalancingOfMargin = balancing
}

func (cf *Configs) GetPercentsToStopSettingNewOrder() items_types.PricePercentType {
	return cf.PercentsToStopSettingNewOrder
}

func (cf *Configs) SetPercentsToStopSettingNewOrder(percent items_types.PricePercentType) {
	cf.PercentsToStopSettingNewOrder = percent
}

func (cf *Configs) GetPercentToDecreasePosition() items_types.PricePercentType {
	return cf.PercentToDecreasePosition
}

func (cf *Configs) SetPercentToDecreasePosition(percent items_types.PricePercentType) {
	cf.PercentToDecreasePosition = percent
}

func (cf *Configs) GetObserverTimeOutMillisecond() int {
	return cf.ObserverTimeOutMillisecond
}

func (cf *Configs) SetObserverTimeOutMillisecond(timeout int) {
	cf.ObserverTimeOutMillisecond = timeout
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
		Connection                *connection_types.Connection `json:"connection"`
		LogLevel                  string                       `json:"log_level"`
		ObservePriceLiquidation   bool                         `json:"observe_price_liquidation"`
		ObservePosition           bool                         `json:"observe_position"`
		RestartClosedPosition     bool                         `json:"close_position_on_restart"`
		BalancingOfMargin         bool                         `json:"balancing_of_margin"`
		PercentsToLiquidation     items_types.PricePercentType `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition items_types.PricePercentType `json:"percent_to_decrease_position"`
		ObserverTimeOut           int                          `json:"observer_timeout_millisecond"`
		Pairs                     []*pairs_types.Pairs         `json:"pairs"`
	}{
		Connection:                c.Connection,
		LogLevel:                  c.LogLevel.String(),
		ObservePriceLiquidation:   c.ObservePriceLiquidation,
		ObservePosition:           c.ObservePosition,
		RestartClosedPosition:     c.ClosePositionOnRestart,
		BalancingOfMargin:         c.BalancingOfMargin,
		PercentsToLiquidation:     c.PercentsToStopSettingNewOrder,
		PercentToDecreasePosition: c.PercentToDecreasePosition,
		ObserverTimeOut:           c.ObserverTimeOutMillisecond,
		Pairs:                     pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		Connection                *connection_types.Connection `json:"connection"`
		LogLevel                  string                       `json:"log_level"`
		ObservePriceLiquidation   bool                         `json:"observe_price_liquidation"`
		ObservePosition           bool                         `json:"observe_position"`
		RestartClosedPosition     bool                         `json:"close_position_on_restart"`
		BalancingOfMargin         bool                         `json:"balancing_of_margin"`
		PercentsToLiquidation     items_types.PricePercentType `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition items_types.PricePercentType `json:"percent_to_decrease_position"`
		ObserverTimeOut           int                          `json:"observer_timeout_millisecond"`
		Pairs                     []*pairs_types.Pairs         `json:"pairs"`
	}{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}
	c.Connection = &connection_types.Connection{
		APIKey:     temp.Connection.APIKey,
		APISecret:  temp.Connection.APISecret,
		UseTestNet: temp.Connection.UseTestNet,
	}
	// Parse the string log level to a logrus.Level
	var err error
	c.LogLevel, err = logrus.ParseLevel(temp.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", temp.LogLevel)
	}
	c.ObservePriceLiquidation = temp.ObservePriceLiquidation
	c.ObservePosition = temp.ObservePosition
	c.ClosePositionOnRestart = temp.RestartClosedPosition
	c.BalancingOfMargin = temp.BalancingOfMargin
	c.PercentsToStopSettingNewOrder = temp.PercentsToLiquidation
	c.PercentToDecreasePosition = temp.PercentToDecreasePosition
	c.ObserverTimeOutMillisecond = temp.ObserverTimeOut
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
		ObservePriceLiquidation:       false,
		ObservePosition:               false,
		ClosePositionOnRestart:        false,
		PercentsToStopSettingNewOrder: 0.05, // 5%
		PercentToDecreasePosition:     0.03, // 3%
		ObserverTimeOutMillisecond:    1000,
		Pairs:                         btree.New(2),
	}
}
