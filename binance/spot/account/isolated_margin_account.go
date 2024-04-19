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
	IsolatedMarginAsset   binance.IsolatedMarginAsset
	IsolatedMarginAccount struct {
		// client                *binance.Client
		isolatedMarginAccount *binance.IsolatedMarginAccount
		assets                *btree.BTree
		mu                    sync.Mutex
		assetsRestriction     map[string]bool
	}
)

func (a *IsolatedMarginAsset) Less(item btree.Item) bool {
	return a.Symbol < item.(*IsolatedMarginAsset).Symbol
}

func (a *IsolatedMarginAsset) Equal(item btree.Item) bool {
	return a.Symbol == item.(*IsolatedMarginAsset).Symbol
}

func (a *IsolatedMarginAccount) GetFreeAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *IsolatedMarginAccount) GetLockedAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *IsolatedMarginAccount) GetTotalAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free) + utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *IsolatedMarginAccount) GetAssets() *btree.BTree {
	return a.assets
}

// ReplaceOrInsert for assets
func (a *IsolatedMarginAccount) AssetUpdate(item Asset) {
	a.assets.ReplaceOrInsert(&item)
}

func NewIsolatedMargin(client *binance.Client, symbols []string) (al *IsolatedMarginAccount, err error) {
	al = &IsolatedMarginAccount{
		isolatedMarginAccount: nil,
		assets:                btree.New(2),
		mu:                    sync.Mutex{},
		assetsRestriction:     make(map[string]bool), // Add the missing field "mapSymbols"
	}
	al.isolatedMarginAccount, err = client.NewGetIsolatedMarginAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, assets := range al.isolatedMarginAccount.Assets {
		if _, exists := al.assetsRestriction[assets.Symbol]; exists || len(al.assetsRestriction) == 0 {
			val := IsolatedMarginAsset(assets)
			al.assets.ReplaceOrInsert(&val)
		}
	}
	return
}
