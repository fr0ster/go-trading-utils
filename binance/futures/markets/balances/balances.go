package balances

import (
	futures_account "github.com/fr0ster/go-trading-utils/binance/futures/account"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

func Init(balances *balances_types.BalanceBTree, futuresBalances *btree.BTree) {
	if futuresBalances == nil {
		return
	}
	futuresBalances.Ascend(func(i btree.Item) bool {
		balance := i.(*futures_account.AccountAsset)
		balances.ReplaceOrInsert(&balances_types.BalanceItemType{
			Asset:  balances_types.AssetType(balance.Asset),
			Free:   utils.ConvStrToFloat64(balance.AvailableBalance),
			Locked: utils.ConvStrToFloat64(balance.WalletBalance) - utils.ConvStrToFloat64(balance.AvailableBalance),
		})
		return true
	})
}
