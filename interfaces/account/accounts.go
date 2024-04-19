package account

import "github.com/google/btree"

type (
	Accounts interface {
		GetFreeAsset(symbol string) (float64, error)
		GetLockedAsset(symbol string) (float64, error)
		GetTotalAsset(symbol string) (float64, error)
		GetAssets() *btree.BTree
	}
)
