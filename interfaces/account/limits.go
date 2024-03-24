package account

type (
	AccountLimits interface {
		GetQuantityLimits() []QuantityLimit
		GetQuantity(limit string) (float64, error)
		GetBalance(symbol string) (float64, error)
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
