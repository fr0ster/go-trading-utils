package balances

import (
	spot_account "github.com/fr0ster/go-trading-utils/binance/spot/account"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

func Init(balances *balances_types.BalanceBTree, spotBalances *btree.BTree) {
	if spotBalances == nil {
		return
	}
	spotBalances.Ascend(func(i btree.Item) bool {
		balance := i.(*spot_account.Balance)
		balances.ReplaceOrInsert(balances_types.BalanceItemType{
			Asset:  balances_types.AssetType(balance.Asset),
			Free:   utils.ConvStrToFloat64(balance.Free),
			Locked: utils.ConvStrToFloat64(balance.Locked),
		})
		return true
	})
}
