package account

import (
	"context"
	"fmt"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
)

type (
	UserAsset     binance.UserAsset
	MarginAccount struct {
		client        *binance.Client
		MarginAccount *binance.MarginAccount
		assets        *btree.BTree
		mu            sync.Mutex
		symbols       map[string]bool
	}
)

func (a *UserAsset) Less(item btree.Item) bool {
	return a.Asset < item.(*UserAsset).Asset
}

func (a *UserAsset) Equal(item btree.Item) bool {
	return a.Asset == item.(*UserAsset).Asset
}

func (a *MarginAccount) GetAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *MarginAccount) GetLockedAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *MarginAccount) GetTotalAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free) + utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *MarginAccount) GetBalances() *btree.BTree {
	return a.assets
}

func (a *MarginAccount) Update() (err error) {
	a.MarginAccount, err = a.client.NewGetMarginAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, balance := range a.MarginAccount.UserAssets {
		if _, exists := a.symbols[balance.Asset]; exists || len(a.symbols) == 0 {
			val := UserAsset(balance)
			a.assets.ReplaceOrInsert(&val)
		}
	}
	return nil
}

func NewMargin(client *binance.Client, symbols []string) (al *MarginAccount, err error) {
	al = &MarginAccount{
		client:        client,
		MarginAccount: nil,
		assets:        btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	err = al.Update()
	return
}
