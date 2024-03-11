package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/google/btree"
	"github.com/sirupsen/logrus"
)

type (
	BalanceItemType struct {
		Asset  string
		Free   float64
		Locked float64
	}
)

var (
	balancesTree = btree.New(2)
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b BalanceItemType) Less(than btree.Item) bool {
	return b.Asset < than.(BalanceItemType).Asset
}

func InitBalancesTree(client *binance.Client) (err error) {
	balances, err := GetBalances(client)
	if err != nil {
		return
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

func GetBalancesTree() *btree.BTree {
	return balancesTree
}

func SetBalancesTree(tree *btree.BTree) (err error) {
	balancesTree = tree
	return
}

func GetBalance(asset string) (res BalanceItemType, err error) {
	item := BalanceItemType{
		Asset: asset,
	}
	item = balancesTree.Get(item).(BalanceItemType)
	return item, nil
}

func GetAccountInfo(client *binance.Client) (res *binance.Account, err error) {
	res, err = client.NewGetAccountService().Do(context.Background())
	return
}

func GetBalances(client *binance.Client) (res []binance.Balance, err error) {
	accountInfo, err := GetAccountInfo(client)
	if err != nil {
		return
	}
	res = accountInfo.Balances
	return
}

func ShowBalancesTree() {
	balancesTree.Ascend(func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}

func ShowBalancesTreeByAsset(symbol SymbolType) {
	balancesTree.AscendGreaterOrEqual(BalanceItemType{Asset: string(symbol)}, func(i btree.Item) bool {
		balance := i.(BalanceItemType)
		logrus.Infof("%s: Free: %f, Locked: %f", balance.Asset, balance.Free, balance.Locked)
		return true
	})
}
