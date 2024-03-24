package account

type (
	AccountLimits interface {
		GetQuantities() []QuantityLimit
		GetQuantity(symbol string) (float64, error)
		GetBalance(symbol string) (float64, error)
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
