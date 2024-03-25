package account

import (
	"errors"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	futuresAccount "github.com/fr0ster/go-trading-utils/binance/futures/markets/account"
	"github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	AccountAsset  futures.AccountAsset
	AccountLimits struct {
		client        *futures.Client
		account       *futuresAccount.AccountType
		accountAssets *btree.BTree
		mu            sync.Mutex
		symbols       map[string]bool
	}
)

func (a *AccountAsset) Less(item btree.Item) bool {
	return a.Asset < item.(*AccountAsset).Asset
}

func (a *AccountAsset) Equals(item btree.Item) bool {
	return a.Asset == item.(*AccountAsset).Asset
}

// GetQuantityLimits implements account.AccountLimits.
func (a *AccountLimits) GetQuantities() (res []account.QuantityLimit) {
	a.accountAssets.Ascend(func(item btree.Item) bool {
		val, _ := Binance2AccountAsset(item)
		if _, exists := a.symbols[val.Asset]; exists || len(a.symbols) == 0 {
			symbolBalance, _ := Binance2AccountAsset(item)
			res = append(res, account.QuantityLimit{Symbol: val.Asset, MaxQty: utils.ConvStrToFloat64(symbolBalance.AvailableBalance)})
		}
		return true
	})
	return
}

func (a *AccountLimits) GetAsset(asset string) (float64, error) {
	item := a.accountAssets.Get(&AccountAsset{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, _ := Binance2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.AvailableBalance), nil
	}
}

func (a *AccountLimits) Update() error {
	for _, asset := range a.account.GetAccountInfo().Assets {
		if _, exists := a.symbols[asset.Asset]; exists || len(a.symbols) == 0 {
			val, _ := Binance2AccountAsset(asset)
			a.accountAssets.ReplaceOrInsert(val)
		}
	}
	return nil
}

func NewAccountLimits(client *futures.Client, symbols []string) (al *AccountLimits) {
	var err error
	al = &AccountLimits{
		client:        client,
		account:       nil,
		accountAssets: btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	al.account, err = futuresAccount.New(al.client, 3)
	if err != nil {
		utils.HandleErr(err)
	}
	for _, asset := range al.account.GetAccountInfo().Assets {
		if _, exists := al.symbols[asset.Asset]; exists || len(al.symbols) == 0 {
			val, _ := Binance2AccountAsset(asset)
			al.accountAssets.ReplaceOrInsert(val)
		}
	}
	return
}

func Binance2AccountAsset(binanceAccountAsset interface{}) (*AccountAsset, error) {
	var accountAsset AccountAsset
	err := copier.Copy(&accountAsset, binanceAccountAsset)
	if err != nil {
		return nil, err
	}
	return &accountAsset, nil
}
