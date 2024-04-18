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
		client                *binance.Client
		isolatedMarginAccount *binance.IsolatedMarginAccount
		assets                *btree.BTree
		mu                    sync.Mutex
		symbols               map[string]bool
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

func (a *IsolatedMarginAccount) GetBalances() *btree.BTree {
	return a.assets
}

// ReplaceOrInsert for assets
func (a *IsolatedMarginAccount) AssetUpdate(item binance.Balance) {
	val := Balance(item)
	a.assets.ReplaceOrInsert(&val)
}

func (a *IsolatedMarginAccount) Update() (err error) {
	a.isolatedMarginAccount, err = a.client.NewGetIsolatedMarginAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, assets := range a.isolatedMarginAccount.Assets {
		if _, exists := a.symbols[assets.Symbol]; exists || len(a.symbols) == 0 {
			val := IsolatedMarginAsset(assets)
			a.assets.ReplaceOrInsert(&val)
		}
	}
	return nil
}

func NewIsolatedMargin(client *binance.Client, symbols []string) (al *IsolatedMarginAccount, err error) {
	al = &IsolatedMarginAccount{
		client:                client,
		isolatedMarginAccount: nil,
		assets:                btree.New(2),
		mu:                    sync.Mutex{},
		symbols:               make(map[string]bool), // Add the missing field "mapSymbols"
	}
	err = al.Update()
	return
}
