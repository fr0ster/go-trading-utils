package balances

import (
	"errors"
	"sync"

	"github.com/google/btree"
	"github.com/sirupsen/logrus"
)

type (
	BalanceBTree struct {
		*btree.BTree
		sync.Mutex
	}
	BalanceItemType struct {
		Asset  string
		Free   float64
		Locked float64
	}
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b *BalanceItemType) Less(than btree.Item) bool {
	return b.Asset < than.(*BalanceItemType).Asset
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

func (tree *BalanceBTree) GetItem(asset string) (res *BalanceItemType, err error) {
	val := &BalanceItemType{
		Asset: asset,
	}
	item := tree.Get(val)
	if item == nil {
		return res, errors.New("item not found")
	}
	return item.(*BalanceItemType), nil
}

func (tree *BalanceBTree) SetItem(item *BalanceItemType) error {
	if item == nil {
		return errors.New("item is nil")
	}
	tree.ReplaceOrInsert(item)
	return nil
}

func (tree *BalanceBTree) Show() {
	tree.Ascend(func(i btree.Item) bool {
		balance := i.(*BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}

func (tree *BalanceBTree) ShowByAsset(asset string) {
	tree.AscendGreaterOrEqual(&BalanceItemType{Asset: asset}, func(i btree.Item) bool {
		balance := i.(*BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return false
	})
}

func (tree *BalanceBTree) Update(spotBalances *btree.BTree) {
	if spotBalances == nil {
		return
	}
	spotBalances.Ascend(func(i btree.Item) bool {
		tree.ReplaceOrInsert(i)
		return true
	})
}

func New(degree int) *BalanceBTree {
	balancesTree := &BalanceBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
	return balancesTree
}
