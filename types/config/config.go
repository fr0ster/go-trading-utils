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
		ObservePositionLoss           bool                         `json:"observe_position_loss"`
		ClosePositionOnRestart        bool                         `json:"close_position_on_restart"`
		BalancingOfMargin             bool                         `json:"balancing_of_margin"`
		PercentsToStopSettingNewOrder float64                      `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition     float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOutMillisecond    int                          `json:"observer_timeout_millisecond"`
		UsingBreakEvenPrice           bool                         `json:"using_break_even_price"`
		DynamicDelta                  bool                         `json:"dynamic_delta"`
		BuyDeltaLoss                  float64                      `json:"buy_delta_loss"`
		SellDeltaLoss                 float64                      `json:"sell_delta_loss"`
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

func (cf *Configs) SetReloadConfig(reload bool) {
	cf.ReloadConfig = reload
}

func (cf *Configs) GetObservePriceLiquidation() bool {
	return cf.ObservePriceLiquidation
}

func (cf *Configs) SetObservePriceLiquidation(observe bool) {
	cf.ObservePriceLiquidation = observe
}

func (cf *Configs) GetObservePositionLoss() bool {
	return cf.ObservePositionLoss
}

func (cf *Configs) SetObservePositionLoss(observe bool) {
	cf.ObservePositionLoss = observe
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

func (cf *Configs) GetPercentsToStopSettingNewOrder() float64 {
	return cf.PercentsToStopSettingNewOrder
}

func (cf *Configs) SetPercentsToStopSettingNewOrder(percent float64) {
	cf.PercentsToStopSettingNewOrder = percent
}

func (cf *Configs) GetPercentToDecreasePosition() float64 {
	return cf.PercentToDecreasePosition
}

func (cf *Configs) SetPercentToDecreasePosition(percent float64) {
	cf.PercentToDecreasePosition = percent
}

func (cf *Configs) GetObserverTimeOutMillisecond() int {
	return cf.ObserverTimeOutMillisecond
}

func (cf *Configs) SetObserverTimeOutMillisecond(timeout int) {
	cf.ObserverTimeOutMillisecond = timeout
}

func (cf *Configs) GetUsingBreakEvenPrice() bool {
	return cf.UsingBreakEvenPrice
}

func (cf *Configs) SetUsingBreakEvenPrice(use bool) {
	cf.UsingBreakEvenPrice = use
}

func (cf *Configs) GetDynamicDelta() bool {
	return cf.DynamicDelta
}

func (cf *Configs) SetDynamicDelta(dynamic bool) {
	cf.DynamicDelta = dynamic
}

func (cf *Configs) GetBuyDeltaLoss() float64 {
	return cf.BuyDeltaLoss
}

func (cf *Configs) SetBuyDeltaLoss(delta float64) {
	cf.BuyDeltaLoss = delta
}

func (cf *Configs) GetSellDeltaLoss() float64 {
	return cf.SellDeltaLoss
}

func (cf *Configs) SetSellDeltaLoss(delta float64) {
	cf.SellDeltaLoss = delta
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
		ReloadConfig              bool                         `json:"reload_config"`
		ObservePriceLiquidation   bool                         `json:"observe_price_liquidation"`
		ObservePositionLoss       bool                         `json:"observe_position_loss"`
		RestartClosedPosition     bool                         `json:"close_position_on_restart"`
		BalancingOfMargin         bool                         `json:"balancing_of_margin"`
		PercentsToLiquidation     float64                      `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOut           int                          `json:"observer_timeout_millisecond"`
		UsingBreakEvenPrice       bool                         `json:"using_break_even_price"`
		DynamicDelta              bool                         `json:"dynamic_delta"`
		BuyDeltaLoss              float64                      `json:"buy_delta_loss"`
		SellDeltaLoss             float64                      `json:"sell_delta_loss"`
		Pairs                     []*pairs_types.Pairs         `json:"pairs"`
	}{
		Connection:                c.Connection,
		LogLevel:                  c.LogLevel.String(),
		ReloadConfig:              c.ReloadConfig,
		ObservePriceLiquidation:   c.ObservePriceLiquidation,
		ObservePositionLoss:       c.ObservePositionLoss,
		RestartClosedPosition:     c.ClosePositionOnRestart,
		BalancingOfMargin:         c.BalancingOfMargin,
		PercentsToLiquidation:     c.PercentsToStopSettingNewOrder,
		PercentToDecreasePosition: c.PercentToDecreasePosition,
		ObserverTimeOut:           c.ObserverTimeOutMillisecond,
		UsingBreakEvenPrice:       c.UsingBreakEvenPrice,
		DynamicDelta:              c.DynamicDelta,
		BuyDeltaLoss:              c.BuyDeltaLoss,
		SellDeltaLoss:             c.SellDeltaLoss,
		Pairs:                     pairs,
	}, "", "  ")
}

func (c *Configs) UnmarshalJSON(data []byte) error {
	temp := &struct {
		Connection                *connection_types.Connection `json:"connection"`
		LogLevel                  string                       `json:"log_level"`
		ReloadConfig              bool                         `json:"reload_config"`
		ObservePriceLiquidation   bool                         `json:"observe_price_liquidation"`
		ObservePositionLoss       bool                         `json:"observe_position_loss"`
		RestartClosedPosition     bool                         `json:"close_position_on_restart"`
		BalancingOfMargin         bool                         `json:"balancing_of_margin"`
		PercentsToLiquidation     float64                      `json:"percents_to_stop_setting_new_order"`
		PercentToDecreasePosition float64                      `json:"percent_to_decrease_position"`
		ObserverTimeOut           int                          `json:"observer_timeout_millisecond"`
		UsingBreakEvenPrice       bool                         `json:"using_break_even_price"`
		DynamicDelta              bool                         `json:"dynamic_delta"`
		BuyDeltaLoss              float64                      `json:"buy_delta_loss"`
		SellDeltaLoss             float64                      `json:"sell_delta_loss"`
		Pairs                     []*pairs_types.Pairs         `json:"pairs"`
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
	c.ObservePositionLoss = temp.ObservePositionLoss
	c.ClosePositionOnRestart = temp.RestartClosedPosition
	c.BalancingOfMargin = temp.BalancingOfMargin
	c.PercentsToStopSettingNewOrder = temp.PercentsToLiquidation
	c.PercentToDecreasePosition = temp.PercentToDecreasePosition
	c.ObserverTimeOutMillisecond = temp.ObserverTimeOut
	c.UsingBreakEvenPrice = temp.UsingBreakEvenPrice
	c.DynamicDelta = temp.DynamicDelta
	c.BuyDeltaLoss = temp.BuyDeltaLoss
	c.SellDeltaLoss = temp.SellDeltaLoss
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
		ObservePositionLoss:           false,
		ClosePositionOnRestart:        false,
		PercentsToStopSettingNewOrder: 0.05,
		PercentToDecreasePosition:     0.03,
		ObserverTimeOutMillisecond:    1000,
		UsingBreakEvenPrice:           false,
		DynamicDelta:                  false,
		BuyDeltaLoss:                  0.015,
		SellDeltaLoss:                 0.015,
		Pairs:                         btree.New(2),
	}
}
