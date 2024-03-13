package markets

import (
	"errors"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
)

type (
	BalanceBTree struct {
		*btree.BTree
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
	item = balancesTree.Get(item).(BalanceItemType)
	return item, nil
}

func (balancesTree *BalanceBTree) SetItem(item BalanceItemType) {
	balancesTree.ReplaceOrInsert(item)
}

func (balancesTree *BalanceBTree) Show() {
	balancesTree.Ascend(func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}

func (balancesTree *DepthBTree) ShowBalancesTreeByAsset(symbol SymbolType) {
	balancesTree.AscendGreaterOrEqual(BalanceItemType{Asset: string(symbol)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return false
	})
}
