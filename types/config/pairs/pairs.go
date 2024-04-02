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
		BuyQuantity  float64     `json:"buy_quantity"`
		BuyValue     float64     `json:"buy_value"`
		SellQuantity float64     `json:"sell_quantity"`
		SellValue    float64     `json:"sell_value"`
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

func (cr *Pairs) GetBuyQuantity() float64 {
	return cr.BuyQuantity
}

func (cr *Pairs) GetSellQuantity() float64 {
	return cr.SellQuantity
}

func (cr *Pairs) GetBuyValue() float64 {
	return cr.BuyValue
}

func (cr *Pairs) GetSellValue() float64 {
	return cr.SellValue
}

func (cr *Pairs) SetBuyQuantity(quantity float64) {
	cr.BuyQuantity = quantity
}

func (cr *Pairs) SetSellQuantity(quantity float64) {
	cr.SellQuantity = quantity
}

func (cr *Pairs) SetBuyValue(value float64) {
	cr.BuyValue = value
}

func (cr *Pairs) SetSellValue(value float64) {
	cr.SellValue = value
}

func (cr *Pairs) GetMiddlePrice() float64 {
	if cr.BuyQuantity == 0 && cr.SellQuantity == 0 {
		return 0
	}

	return (cr.BuyValue - cr.SellValue) / (cr.BuyQuantity - cr.SellQuantity)
}
