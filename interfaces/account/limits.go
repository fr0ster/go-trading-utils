package account

type (
	AccountLimits interface {
		GetQuantityLimits() []QuantityLimit
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
