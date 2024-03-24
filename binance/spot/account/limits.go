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
func (a *AccountLimits) GetQuantityLimits() (res []account.QuantityLimit) {
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

func (a *AccountLimits) getValue(asset string) (float64, error) {
	item := a.assetBalances.Get(&Balance{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, _ := Binance2AssetBalance(item)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

// GetBalance implements account.AccountLimits.
func (a *AccountLimits) GetBalance(symbol string) (res float64, err error) {
	return a.getValue(symbol)
}

// GetQuantity implements account.AccountLimits.
func (a *AccountLimits) GetQuantity(symbol string) (float64, error) {
	return a.getValue(symbol)
}

func NewAccountLimits(client *binance.Client, symbols []string) (al *AccountLimits) {
	al = &AccountLimits{
		assetBalances: btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	spotAccount, err := spotAccount.New(client, 3)
	if err != nil {
		utils.HandleErr(err)
	}
	for _, balance := range spotAccount.GetAccountInfo().Balances {
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
