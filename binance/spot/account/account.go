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
	Asset      binance.Balance
	Permission struct{ string }
	Account    struct {
		AccountUpdateTime int64
		MakerCommission   int64                   `json:"makerCommission"`
		TakerCommission   int64                   `json:"takerCommission"`
		BuyerCommission   int64                   `json:"buyerCommission"`
		SellerCommission  int64                   `json:"sellerCommission"`
		CommissionRates   binance.CommissionRates `json:"commissionRates"`
		CanTrade          bool                    `json:"canTrade"`
		CanWithdraw       bool                    `json:"canWithdraw"`
		CanDeposit        bool                    `json:"canDeposit"`
		UpdateTime        uint64                  `json:"updateTime"`
		AccountType       string                  `json:"accountType"`
		assets            *btree.BTree
		permissions       *btree.BTree
		mu                sync.Mutex
		assetsRestriction map[string]bool
	}
)

func (a *Asset) Less(item btree.Item) bool {
	return a.Asset < item.(*Asset).Asset
}

func (a *Asset) Equal(item btree.Item) bool {
	return a.Asset == item.(*Asset).Asset
}

func (a *Permission) Less(item btree.Item) bool {
	return a.string < item.(*Permission).string
}

func (a *Permission) Equal(item btree.Item) bool {
	return a.string == item.(*Permission).string
}

// Locking the account
func (a *Account) Lock() {
	a.mu.Lock()
}

// Unlocking the account
func (a *Account) Unlock() {
	a.mu.Unlock()
}

// Get MakerCommission
func (a *Account) GetMakerCommission() float64 {
	return float64(a.MakerCommission / 10000)
}

// Get TakerCommission
func (a *Account) GetTakerCommission() float64 {
	return float64(a.TakerCommission / 10000)
}

// Get BuyerCommission
func (a *Account) GetBuyerCommission() float64 {
	return float64(a.BuyerCommission / 10000)
}

// Get SellerCommission
func (a *Account) GetSellerCommission() float64 {
	return float64(a.SellerCommission / 10000)
}

func (a *Account) GetFreeAsset(asset string) (float64, error) {
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Asset)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *Account) GetLockedAsset(asset string) (float64, error) {
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Asset)
		return utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *Account) GetTotalAsset(asset string) (float64, error) {
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Asset)
		return utils.ConvStrToFloat64(symbolBalance.Free) + utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *Account) GetAssets() *btree.BTree {
	return a.assets
}

func (a *Account) GetPermissions() *btree.BTree {
	return a.permissions
}

func (a *Account) AssetsAscend(iterator func(item *Asset) bool) {
	a.assets.Ascend(func(i btree.Item) bool {
		return iterator(i.(*Asset))
	})
}

func (a *Account) PermissionsAscend(iterator func(item *Permission) bool) {
	a.permissions.Ascend(func(i btree.Item) bool {
		return iterator(i.(*Permission))
	})
}

func (a *Account) AssetsDescend(iterator func(item *Asset) bool) {
	a.assets.Descend(func(i btree.Item) bool {
		return iterator(i.(*Asset))
	})
}

func (a *Account) PermissionsDescend(iterator func(item *Permission) bool) {
	a.permissions.Descend(func(i btree.Item) bool {
		return iterator(i.(*Permission))
	})
}

// ReplaceOrInsert for assets
func (a *Account) AssetUpdate(item Asset) {
	a.assets.ReplaceOrInsert(&item)
}

// ReplaceOrInsert for permissions
func (a *Account) PermissionUpdate(item string) {
	a.permissions.ReplaceOrInsert(&Permission{string: item})
}

func New(client *binance.Client, assets []string) (account *Account, err error) {
	accountIn, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	account = &Account{
		AccountUpdateTime: int64(accountIn.UpdateTime),
		MakerCommission:   accountIn.MakerCommission,
		TakerCommission:   accountIn.TakerCommission,
		BuyerCommission:   accountIn.BuyerCommission,
		SellerCommission:  accountIn.SellerCommission,
		CommissionRates:   accountIn.CommissionRates,
		CanTrade:          accountIn.CanTrade,
		CanWithdraw:       accountIn.CanWithdraw,
		CanDeposit:        accountIn.CanDeposit,
		UpdateTime:        accountIn.UpdateTime,
		AccountType:       accountIn.AccountType,
		assets:            btree.New(2),
		permissions:       btree.New(2),
		mu:                sync.Mutex{},
		assetsRestriction: make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, asset := range assets {
		account.assetsRestriction[asset] = true
	}
	for _, balance := range accountIn.Balances {
		if _, exists := account.assetsRestriction[balance.Asset]; exists || len(account.assetsRestriction) == 0 {
			account.AssetUpdate(Asset(balance))
		}
	}
	for _, permission := range accountIn.Permissions {
		account.PermissionUpdate(permission)
	}
	return
}
