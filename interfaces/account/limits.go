package account

import "github.com/google/btree"

type (
	AccountLimits interface {
		GetQuantities() []QuantityLimit
		GetAsset(symbol string) (float64, error)
		Update() error
		GetBalances() *btree.BTree
	}
	QuantityLimit struct {
		Symbol string
		MaxQty float64
	}
)
