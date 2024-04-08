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
		AccountType  AccountType `json:"account_type"`  // Тип акаунта
		Pair         string      `json:"symbol"`        // Пара
		TargetSymbol string      `json:"target_symbol"` // Цільовий токен
		BaseSymbol   string      `json:"base_symbol"`   // Базовий токен

		// Ліміт на вхід в позицію, відсоток від балансу базової валюти,
		// поки не наберемо цей ліміт, не можемо перейти до режиму спекуляціі
		// Режим входу - накопичуємо цільовий токен
		// Режим спекуляції - купуємо/продаемо цільовий токен за базовий
		// Режим виходу - продаемо цільовий токен
		LimitInputIntoPosition float64 `json:"limit_input_into_position"`
		LimitInPosition        float64 `json:"limit_in_position"`    // Ліміт на позицію, відсоток від балансу базової валюти
		LimitOnTransaction     float64 `json:"limit_on_transaction"` // Ліміт на транзакцію, відсоток від ліміту на позицію

		BuyDelta     float64 `json:"buy_delta"`     // Дельта для купівлі
		BuyQuantity  float64 `json:"buy_quantity"`  // Кількість для купівлі, суммарно по позиції
		BuyValue     float64 `json:"buy_value"`     // Вартість для купівлі, суммарно по позиції
		SellDelta    float64 `json:"sell_delta"`    // Дельта для продажу, суммарно по позиції
		SellQuantity float64 `json:"sell_quantity"` // Кількість для продажу, суммарно по позиції
		SellValue    float64 `json:"sell_value"`    // Вартість для продажу, суммарно по позиції
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

func (cr *Pairs) GetLimitInputIntoPosition() float64 {
	return cr.LimitInputIntoPosition
}

func (cr *Pairs) GetLimitInPosition() float64 {
	return cr.LimitInPosition
}

func (cr *Pairs) GetLimitOnTransaction() float64 {
	return cr.LimitOnTransaction
}

func (cr *Pairs) GetBuyDelta() float64 {
	return cr.BuyDelta
}

func (cr *Pairs) GetSellDelta() float64 {
	return cr.SellDelta
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

func (cr *Pairs) GetProfit(currentPrice float64) float64 {
	return (currentPrice - cr.GetMiddlePrice()) * (cr.BuyQuantity - cr.SellQuantity)
}
