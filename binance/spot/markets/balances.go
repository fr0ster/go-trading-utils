package markets

import (
	"sync"

	"github.com/adshao/go-binance/v2"
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
		Asset  AssetType
		Free   float64
		Locked float64
	}
)

func BalanceNew(degree int, balances []binance.Balance) *BalanceBTree {
	balancesTree := &BalanceBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
	for _, balance := range balances {
		balancesTree.ReplaceOrInsert(BalanceItemType{
			Asset:  AssetType(balance.Asset),
			Free:   utils.ConvStrToFloat64(balance.Free),
			Locked: utils.ConvStrToFloat64(balance.Locked),
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

func (tree *BalanceBTree) Lock() {
	tree.Mutex.Lock()
}

func (tree *BalanceBTree) Unlock() {
	tree.Mutex.Unlock()
}

// func (tree *BalanceBTree) Init(balances []binance.Balance) (err error) {
// 	if len(balances) == 0 {
// 		return errors.New("balances is empty")
// 	}
// 	for _, balance := range balances {
// 		tree.ReplaceOrInsert(BalanceItemType{
// 			Asset:  AssetType(balance.Asset),
// 			Free:   utils.ConvStrToFloat64(balance.Free),
// 			Locked: utils.ConvStrToFloat64(balance.Locked),
// 		})
// 	}
// 	return nil
// }

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
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}

func (tree *BalanceBTree) ShowByAsset(symbol SymbolType) {
	tree.AscendGreaterOrEqual(BalanceItemType{Asset: AssetType(symbol)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return false
	})
}
