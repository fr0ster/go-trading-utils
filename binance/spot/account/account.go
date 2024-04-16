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
	Balance    binance.Balance
	Permission struct{ string }
	Account    struct {
		client      *binance.Client
		account     *binance.Account
		balances    *btree.BTree
		permissions *btree.BTree
		mu          sync.Mutex
		symbols     map[string]bool
	}
)

func (a *Balance) Less(item btree.Item) bool {
	return a.Asset < item.(*Balance).Asset
}

func (a *Balance) Equal(item btree.Item) bool {
	return a.Asset == item.(*Balance).Asset
}

func (a *Permission) Less(item btree.Item) bool {
	return a.string < item.(*Permission).string
}

func (a *Permission) Equal(item btree.Item) bool {
	return a.string == item.(*Permission).string
}

func (a *Account) GetAsset(asset string) (float64, error) {
	item := a.balances.Get(&Balance{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Balance)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *Account) GetLockedAsset(asset string) (float64, error) {
	item := a.balances.Get(&Balance{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Balance)
		return utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *Account) GetTotalAsset(asset string) (float64, error) {
	item := a.balances.Get(&Balance{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*Balance)
		return utils.ConvStrToFloat64(symbolBalance.Free) + utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *Account) GetPermissions() []string {
	return a.account.Permissions
}

func (a *Account) GetBalances() *btree.BTree {
	return a.balances
}

func (a *Account) Update() (err error) {
	a.account, err = a.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	for _, balance := range a.account.Balances {
		if _, exists := a.symbols[balance.Asset]; exists || len(a.symbols) == 0 {
			val := Balance(balance)
			a.balances.ReplaceOrInsert(&val)
		}
	}
	for _, permission := range a.account.Permissions {
		if _, exists := a.symbols[permission]; exists || len(a.symbols) == 0 {
			a.permissions.ReplaceOrInsert(&Permission{string: permission})
		}
	}
	return nil
}

func New(client *binance.Client, symbols []string) (al *Account, err error) {
	al = &Account{
		client:      client,
		account:     nil,
		balances:    btree.New(2),
		permissions: btree.New(2),
		mu:          sync.Mutex{},
		symbols:     make(map[string]bool), // Add the missing field "mapSymbols"
	}
	err = al.Update()
	return
}

// func Binance2Balances(binanceAssetBalance interface{}) (*balances_types.BalanceItemType, error) {
// 	var assetBalance balances_types.BalanceItemType

// 	val := reflect.ValueOf(binanceAssetBalance)
// 	if val.Kind() != reflect.Ptr {
// 		val = reflect.ValueOf(&binanceAssetBalance)
// 	}

// 	err := copier.Copy(&assetBalance, val.Interface())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &assetBalance, nil
// }

// func Binance2Balances(binanceAccountAsset interface{}) *Balance {
// 	var accountAsset Balance
// 	accountAsset.Asset = binanceAccountAsset.(*Balance).Asset
// 	accountAsset.Free = utils.ConvStrToFloat64(binanceAccountAsset.(*Balance).Free)
// 	accountAsset.Locked = utils.ConvStrToFloat64(binanceAccountAsset.(*Balance).Locked)
// 	return &accountAsset
// }
