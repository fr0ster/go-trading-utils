package config

import (
	"github.com/google/btree"
)

const (
	// SpotAccountType is a constant for spot account type.
	// SPOT/MARGIN/ISOLATED_MARGIN/USDT_FUTURE/COIN_FUTURE
	SpotAccountType    AccountType = "SPOT"
	MarginAccountType  AccountType = "MARGIN"
	IsolatedMarginType AccountType = "ISOLATED_MARGIN"
	USDTFutureType     AccountType = "USDT_FUTURE"
	CoinFutureType     AccountType = "COIN_FUTURE"
)

type (
	AccountType string
	Pairs       struct {
		AccountType  AccountType `json:"account_type"`
		Pair         string      `json:"symbol"`
		TargetSymbol string      `json:"target_symbol"`
		BaseSymbol   string      `json:"base_symbol"`
		Limit        float64     `json:"limit"`
		Delta        float64     `json:"delta"`
		Quantity     float64     `json:"quantity"`
		Value        float64     `json:"value"`
	}
)

func (cr *Pairs) Less(item btree.Item) bool {
	return cr.Pair < item.(*Pairs).Pair
}

func (cr *Pairs) Equals(item btree.Item) bool {
	return cr.Pair == item.(*Pairs).Pair
}

// Get AccountType implements Configuration.
func (cr *Pairs) GetAccountType() AccountType {
	return cr.AccountType
}

// GetSymbol implements Configuration.
func (cr *Pairs) GetPair() string {
	return cr.Pair
}

// GetBaseSymbol implements config.Configuration.
func (cr *Pairs) GetBaseSymbol() string {
	return cr.BaseSymbol
}

// GetTargetSymbol implements config.Configuration.
func (cr *Pairs) GetTargetSymbol() string {
	return cr.TargetSymbol
}

func (cr *Pairs) GetLimit() float64 {
	return cr.Limit
}

func (cr *Pairs) GetDelta() float64 {
	return cr.Delta
}

func (cr *Pairs) GetQuantity() float64 {
	return cr.Quantity
}

func (cr *Pairs) GetValue() float64 {
	return cr.Value
}

func (cr *Pairs) SetLimit(limit float64) {
	cr.Limit = limit
}

func (cr *Pairs) SetQuantity(quantity float64) {
	cr.Quantity = quantity
}

func (cr *Pairs) SetValue(value float64) {
	cr.Value = value
}
