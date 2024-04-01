package account

import (
	"context"
	"errors"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Account struct {
		client           *futures.Client
		account          *futures.Account
		accountAssets    *btree.BTree
		accountPositions *btree.BTree
		mu               sync.Mutex
		symbols          map[string]bool
	}
)

func (a *Account) GetAsset(asset string) (float64, error) {
	item := a.accountAssets.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance := item.(*balances_types.BalanceItemType).Free
		return symbolBalance, nil
	}
}

func (a *Account) Update() error {
	for _, asset := range a.account.Assets {
		if _, exists := a.symbols[asset.Asset]; exists || len(a.symbols) == 0 {
			val, _ := Binance2AccountAsset(asset)
			a.accountAssets.ReplaceOrInsert(val)
		}
	}
	return nil
}

// GetBalances implements account.AccountLimits.
func (a *Account) GetBalances() *btree.BTree {
	return a.accountAssets
}

func NewAccountLimits(client *futures.Client, symbols []string) (al *Account) {
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}
	al = &Account{
		client:           client,
		account:          account,
		accountAssets:    btree.New(2),
		accountPositions: btree.New(2),
		mu:               sync.Mutex{},
		symbols:          make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	for _, asset := range al.account.Assets {
		if _, exists := al.symbols[asset.Asset]; exists || len(al.symbols) == 0 {
			val, _ := Binance2AccountAsset(asset)
			al.accountAssets.ReplaceOrInsert(val)
		}
	}
	for _, position := range al.account.Positions {
		if _, exists := al.symbols[position.Symbol]; exists || len(al.symbols) == 0 {
			val, _ := Binance2AccountAsset(position)
			al.accountPositions.ReplaceOrInsert(val)
		}
	}
	return
}

func Binance2AccountAsset(binanceAccountAsset interface{}) (*balances_types.BalanceItemType, error) {
	var accountAsset balances_types.BalanceItemType
	err := copier.Copy(&accountAsset, binanceAccountAsset)
	if err != nil {
		return nil, err
	}
	return &accountAsset, nil
}
