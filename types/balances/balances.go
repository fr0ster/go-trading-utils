package balances

import (
	"sync"

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

func (tree *BalanceBTree) ShowByAsset(asset string) {
	tree.AscendGreaterOrEqual(BalanceItemType{Asset: AssetType(asset)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return false
	})
}

func New(degree int) *BalanceBTree {
	balancesTree := &BalanceBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
	return balancesTree
}
