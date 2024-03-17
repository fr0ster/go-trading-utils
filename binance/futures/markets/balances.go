package markets

import (
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
)

type (
	AssetType    string
	BalanceBTree struct {
		*btree.BTree
		sync.Mutex
	}
	BalanceItemType struct {
		Asset              AssetType
		Balance            float64
		CrossWalletBalance float64
		ChangeBalance      float64
	}
)

func BalanceNew(degree int, balances []futures.WsBalance) *BalanceBTree {
	balancesTree := &BalanceBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
	for _, balance := range balances {
		balancesTree.ReplaceOrInsert(BalanceItemType{
			Asset:              AssetType(balance.Asset),
			Balance:            utils.ConvStrToFloat64(balance.Balance),
			CrossWalletBalance: utils.ConvStrToFloat64(balance.CrossWalletBalance),
			ChangeBalance:      utils.ConvStrToFloat64(balance.ChangeBalance),
		})
	}
	return balancesTree
}

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b BalanceItemType) Less(than btree.Item) bool {
	return b.Asset < than.(BalanceItemType).Asset
}

func (b *BalanceItemType) Equal(than btree.Item) bool {
	return b.Asset == than.(*BalanceItemType).Asset
}

func (i *BalanceItemType) GetItem() *BalanceItemType {
	return i
}

func (tree *BalanceBTree) Lock() {
	tree.Mutex.Lock()
}

func (tree *BalanceBTree) Unlock() {
	tree.Mutex.Unlock()
}

func (tree *BalanceBTree) GetItem(asset AssetType) (res BalanceItemType, err error) {
	item := BalanceItemType{
		Asset: AssetType(asset),
	}
	item = tree.Get(item).(BalanceItemType)
	return item, nil
}

func (tree *BalanceBTree) SetItem(item BalanceItemType) {
	tree.ReplaceOrInsert(item)
}

func (tree *BalanceBTree) Show() {
	tree.Ascend(func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof(
			"%s: Balance: %f, CrossWalletBalance: %f, ChangeBalance: %f",
			balance.Asset,
			balance.Balance,
			balance.CrossWalletBalance,
			balance.ChangeBalance)
		return true
	})
}

func (tree *DepthBTree) ShowByAsset(symbol SymbolType) {
	tree.AscendGreaterOrEqual(BalanceItemType{Asset: AssetType(symbol)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof(
			"%s: Balance: %f, CrossWalletBalance: %f, ChangeBalance: %f",
			balance.Asset,
			balance.Balance,
			balance.CrossWalletBalance,
			balance.ChangeBalance)
		return false
	})
}
