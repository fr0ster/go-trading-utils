package balances

import (
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
)

type (
	Balances interface {
		Lock()
		Unlock()
		GetItem(asset string) (*balances_types.BalanceItemType, error)
		SetItem(item *balances_types.BalanceItemType) error
		Show()
		ShowByAsset(asset string)
	}
)
