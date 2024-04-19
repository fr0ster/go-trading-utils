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
		client            *binance.Client
		marginAccount     *binance.MarginAccount
		assets            *btree.BTree
		mu                sync.Mutex
		assetsRestriction map[string]bool
	}
)

func (a *UserAsset) Less(item btree.Item) bool {
	return a.Asset < item.(*UserAsset).Asset
}

func (a *UserAsset) Equal(item btree.Item) bool {
	return a.Asset == item.(*UserAsset).Asset
}

func (a *MarginAccount) GetFreeAsset(asset string) (float64, error) {
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

func (a *MarginAccount) GetAssets() *btree.BTree {
	return a.assets
}

// ReplaceOrInsert for assets
func (a *MarginAccount) AssetUpdate(item binance.Balance) {
	val := Asset(item)
	a.assets.ReplaceOrInsert(&val)
}

func NewMargin(client *binance.Client, symbols []string) (al *MarginAccount, err error) {
	al = &MarginAccount{
		client:            client,
		marginAccount:     nil,
		assets:            btree.New(2),
		mu:                sync.Mutex{},
		assetsRestriction: make(map[string]bool), // Add the missing field "mapSymbols"
	}

	al.marginAccount, err = client.NewGetMarginAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, balance := range al.marginAccount.UserAssets {
		if _, exists := al.assetsRestriction[balance.Asset]; exists || len(al.assetsRestriction) == 0 {
			val := UserAsset(balance)
			al.assets.ReplaceOrInsert(&val)
		}
	}
	return
}
