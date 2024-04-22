package account

import "github.com/google/btree"

type (
	Accounts interface {
		Lock()
		Unlock()
		GetTakerCommission() float64
		GetMakerCommission() float64
		GetBuyerCommission() float64
		GetSellerCommission() float64
		GetFreeAsset(symbol string) (float64, error)
		GetLockedAsset(symbol string) (float64, error)
		GetTotalAsset(symbol string) (float64, error)
		GetAssets() *btree.BTree
	}
)
