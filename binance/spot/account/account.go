package account

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/adshao/go-binance/v2"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Account struct {
		client        *binance.Client
		account       *binance.Account
		assetBalances *btree.BTree
		mu            sync.Mutex
		symbols       map[string]bool
	}
)

func (a *Account) GetAsset(asset string) (float64, error) {
	item := a.assetBalances.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance, _ := Binance2AssetBalance(item)
		return symbolBalance.Free, nil
	}
}

func (a *Account) GetLockedAsset(asset string) (float64, error) {
	item := a.assetBalances.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance, _ := Binance2AssetBalance(item)
		return symbolBalance.Locked, nil
	}
}

func (a *Account) GetTotalAsset(asset string) (float64, error) {
	item := a.assetBalances.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance, _ := Binance2AssetBalance(item)
		return symbolBalance.Free + symbolBalance.Locked, nil
	}
}

func (a *Account) GetPermissions() []string {
	return a.account.Permissions
}

func (a *Account) Update() (err error) {
	a.account, err = a.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, balance := range a.account.Balances {
		if _, exists := a.symbols[balance.Asset]; exists || len(a.symbols) == 0 {
			val := balances_types.BalanceItemType{
				Asset:  balance.Asset,
				Free:   utils.ConvStrToFloat64(balance.Free),
				Locked: utils.ConvStrToFloat64(balance.Locked),
			}
			a.assetBalances.ReplaceOrInsert(&val)
		}
	}
	return nil
}

func (a *Account) GetBalances() *btree.BTree {
	return a.assetBalances
}

func New(client *binance.Client, symbols []string) (al *Account, err error) {
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	al = &Account{
		client:        client,
		account:       account,
		assetBalances: btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	for _, balance := range al.account.Balances {
		if _, exists := al.symbols[balance.Asset]; exists || len(al.symbols) == 0 {
			val := balances_types.BalanceItemType{
				Asset:  balance.Asset,
				Free:   utils.ConvStrToFloat64(balance.Free),
				Locked: utils.ConvStrToFloat64(balance.Locked),
			}
			al.assetBalances.ReplaceOrInsert(&val)
		}
	}
	return
}

func Binance2AssetBalance(binanceAssetBalance interface{}) (*balances_types.BalanceItemType, error) {
	var assetBalance balances_types.BalanceItemType

	val := reflect.ValueOf(binanceAssetBalance)
	if val.Kind() != reflect.Ptr {
		val = reflect.ValueOf(&binanceAssetBalance)
	}

	err := copier.Copy(&assetBalance, val.Interface())
	if err != nil {
		return nil, err
	}

	return &assetBalance, nil
}
