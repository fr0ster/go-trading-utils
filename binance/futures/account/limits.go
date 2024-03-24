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
		// binance.AssetBalance
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
func (a *AccountLimits) GetQuantityLimits() (res []account.QuantityLimit) {
	for symbol := range a.symbols {
		item, _ := a.getValue(symbol)
		symbolBalance, _ := Binance2AccountAsset(item)
		res = append(res, account.QuantityLimit{Symbol: symbol, MaxQty: utils.ConvStrToFloat64(symbolBalance.AvailableBalance)})
	}
	return
}

func (a *AccountLimits) getValue(asset string) (float64, error) {
	item := a.accountAssets.Get(&AccountAsset{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, _ := Binance2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.AvailableBalance), nil
	}
}

// GetBalance implements account.AccountLimits.
func (a *AccountLimits) GetBalance(asset string) (res float64, err error) {
	return a.getValue(asset)
}

// GetQuantity implements account.AccountLimits.
func (a *AccountLimits) GetQuantity(symbol string) (float64, error) {
	return a.getValue(symbol)
}

func NewAccountLimits(client *futures.Client, symbols []string) (al *AccountLimits) {
	al = &AccountLimits{
		accountAssets: btree.New(2),
		mu:            sync.Mutex{},
		symbols:       make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, symbol := range symbols {
		al.symbols[symbol] = true
	}
	spotAccount, err := futuresAccount.New(client, 3)
	if err != nil {
		utils.HandleErr(err)
	}
	for _, asset := range spotAccount.GetAccountInfo().Assets {
		if _, exists := al.symbols[asset.Asset]; exists {
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
