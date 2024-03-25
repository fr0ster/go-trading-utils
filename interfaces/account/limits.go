package account

type (
	AccountLimits interface {
		GetQuantities() []QuantityLimit
		GetAsset(symbol string) (float64, error)
		Update() error
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
