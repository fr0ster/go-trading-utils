package account

import (
	"errors"
	"reflect"
	"sync"

	"github.com/adshao/go-binance/v2"
	spotAccount "github.com/fr0ster/go-trading-utils/binance/spot/markets/account"
	"github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Balance       binance.Balance
	AccountLimits struct {
		client        *binance.Client
		account       *spotAccount.AccountType
		assetBalances *btree.BTree
		mu            sync.Mutex
		symbols       map[string]bool
	}
)

func (a *Balance) Less(item btree.Item) bool {
	return a.Asset < item.(*Balance).Asset
}

func (a *Balance) Equals(item btree.Item) bool {
	return a.Asset == item.(*Balance).Asset
}

// GetQuantityLimits implements account.AccountLimits.
func (a *AccountLimits) GetQuantities() (res []account.QuantityLimit) {
	a.assetBalances.Ascend(func(item btree.Item) bool {
		val, _ := Binance2AssetBalance(item)
		if _, exists := a.symbols[val.Asset]; exists || len(a.symbols) == 0 {
			symbolBalance, _ := Binance2AssetBalance(item)
			res = append(res, account.QuantityLimit{Symbol: val.Asset, MaxQty: utils.ConvStrToFloat64(symbolBalance.Free)})
		}
		return true
	})
	return
}

func (a *AccountLimits) GetAsset(asset string) (float64, error) {
	item := a.assetBalances.Get(&Balance{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, _ := Binance2AssetBalance(item)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *AccountLimits) Update() error {
	for _, balance := range a.account.GetAccountInfo().Balances {
		if _, exists := a.symbols[balance.Asset]; exists || len(a.symbols) == 0 {
			val := Balance(balance)
			a.assetBalances.ReplaceOrInsert(&val)
		}
	}
	return nil
}

func NewAccountLimits(client *binance.Client, symbols []string) (al *AccountLimits) {
	var err error
	al = &AccountLimits{
		client:        client,
		account:       nil,
		assetBalances: btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	al.account, err = spotAccount.New(al.client, 3)
	if err != nil {
		utils.HandleErr(err)
	}
	for _, balance := range al.account.GetAccountInfo().Balances {
		if _, exists := al.symbols[balance.Asset]; exists || len(al.symbols) == 0 {
			val := Balance(balance)
			al.assetBalances.ReplaceOrInsert(&val)
		}
	}
	return
}

func Binance2AssetBalance(binanceAssetBalance interface{}) (*Balance, error) {
	var assetBalance Balance

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
