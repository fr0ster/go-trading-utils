package account

type (
	AccountLimits interface {
		GetQuantities() []QuantityLimit
		GetAsset(symbol string) (float64, error)
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
