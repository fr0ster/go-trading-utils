package account

import "github.com/google/btree"

type (
	Accounts interface {
		GetAsset(symbol string) (float64, error)
		Update() error
		GetBalances() *btree.BTree
	}
)
