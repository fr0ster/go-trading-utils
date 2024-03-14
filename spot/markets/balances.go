package markets

import (
	"errors"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
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

func BalanceNew(degree int) *BalanceBTree {
	return &BalanceBTree{
		BTree: btree.New(degree),
		Mutex: sync.Mutex{},
	}
}

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b BalanceItemType) Less(than btree.Item) bool {
	return b.Asset < than.(BalanceItemType).Asset
}

func (b *BalanceItemType) Equal(than btree.Item) bool {
	return b.Asset == than.(*BalanceItemType).Asset
}

func (balancesTree *BalanceBTree) Init(balances []binance.Balance) (err error) {
	if len(balances) == 0 {
		return errors.New("balances is empty")
	}
	balancesTree.Mutex.Lock()
	defer balancesTree.Mutex.Unlock()
	for _, balance := range balances {
		balancesTree.ReplaceOrInsert(BalanceItemType{
			Asset:  balance.Asset,
			Free:   utils.ConvStrToFloat64(balance.Free),
			Locked: utils.ConvStrToFloat64(balance.Locked),
		})
	}
	return nil
}

func (balancesTree *BalanceBTree) GetItem(asset string) (res BalanceItemType, err error) {
	item := BalanceItemType{
		Asset: asset,
	}
	balancesTree.Mutex.Lock()
	defer balancesTree.Mutex.Unlock()
	item = balancesTree.Get(item).(BalanceItemType)
	return item, nil
}

func (balancesTree *BalanceBTree) SetItem(item BalanceItemType) {
	balancesTree.Mutex.Lock()
	defer balancesTree.Mutex.Unlock()
	balancesTree.ReplaceOrInsert(item)
}

func (balancesTree *BalanceBTree) Show() {
	balancesTree.Mutex.Lock()
	defer balancesTree.Mutex.Unlock()
	balancesTree.Ascend(func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}

func (balancesTree *DepthBTree) ShowByAsset(symbol SymbolType) {
	balancesTree.Mutex.Lock()
	defer balancesTree.Mutex.Unlock()
	balancesTree.AscendGreaterOrEqual(BalanceItemType{Asset: string(symbol)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return false
	})
}
