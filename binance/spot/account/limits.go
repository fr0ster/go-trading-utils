package account

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	spotAccount "github.com/fr0ster/go-trading-utils/binance/spot/markets/account"
	"github.com/fr0ster/go-trading-utils/interfaces/account"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	AssetBalance  binance.AssetBalance
	AccountLimits struct {
		// binance.AssetBalance
		assetBalances *btree.BTree
		mu            sync.Mutex
		symbols       map[string]bool
	}
)

func (a *AssetBalance) Less(item btree.Item) bool {
	return a.Asset < item.(*AssetBalance).Asset
}

func (a *AssetBalance) Equals(item btree.Item) bool {
	return a.Asset == item.(*AssetBalance).Asset
}

// GetQuantityLimits implements account.AccountLimits.
func (a *AccountLimits) GetQuantityLimits() (res []account.QuantityLimit) {
	for symbol := range a.symbols {
		item := a.assetBalances.Get(&AssetBalance{Asset: symbol})
		if item == nil {
			res = append(res, account.QuantityLimit{Symbol: symbol, MaxQty: 0})
		} else {
			symbolBalance, _ := Binance2AssetBalance(item)
			res = append(res, account.QuantityLimit{Symbol: symbol, MaxQty: symbolBalance.Free})
		}
	}
	return
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
		if _, exists := al.symbols[balance.Asset]; exists {
			val, _ := Binance2AssetBalance(balance)
			al.assetBalances.ReplaceOrInsert(val)
		}
	}
	return
}

func Binance2AssetBalance(binanceAssetBalance interface{}) (*AssetBalance, error) {
	var assetBalance AssetBalance
	err := copier.Copy(&assetBalance, binanceAssetBalance)
	if err != nil {
		return nil, err
	}
	return &assetBalance, nil
}
