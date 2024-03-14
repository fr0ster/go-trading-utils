package spot

import (
	"time"

	"github.com/adshao/go-binance/v2"
)

type Config struct {
	Timestamp         time.Time
	AccountType       binance.AccountType
	Symbol            binance.SymbolType
	Balance           float64
	CalculatedBalance float64
	Quantity          float64
	Value             float64
	BoundQuantity     float64
}

type Log struct {
	Timestamp time.Time
	Message   string
}
