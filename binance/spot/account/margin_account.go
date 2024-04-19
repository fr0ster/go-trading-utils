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
	UserAsset     binance.UserAsset
	MarginAccount struct {
		BorrowEnabled       bool   `json:"borrowEnabled"`
		MarginLevel         string `json:"marginLevel"`
		TotalAssetOfBTC     string `json:"totalAssetOfBtc"`
		TotalLiabilityOfBTC string `json:"totalLiabilityOfBtc"`
		TotalNetAssetOfBTC  string `json:"totalNetAssetOfBtc"`
		TradeEnabled        bool   `json:"tradeEnabled"`
		TransferEnabled     bool   `json:"transferEnabled"`
		assets              *btree.BTree
		mu                  sync.Mutex
		assetsRestriction   map[string]bool
	}
)

func (a *UserAsset) Less(item btree.Item) bool {
	return a.Asset < item.(*UserAsset).Asset
}

func (a *UserAsset) Equal(item btree.Item) bool {
	return a.Asset == item.(*UserAsset).Asset
}

func (a *MarginAccount) GetFreeAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free), nil
	}
}

func (a *MarginAccount) GetLockedAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *MarginAccount) GetTotalAsset(asset string) (float64, error) {
	item := a.assets.Get(&UserAsset{Asset: asset})
	if item == nil {
		return 0, fmt.Errorf("%s not found", asset)
	} else {
		symbolBalance := item.(*UserAsset)
		return utils.ConvStrToFloat64(symbolBalance.Free) + utils.ConvStrToFloat64(symbolBalance.Locked), nil
	}
}

func (a *MarginAccount) GetAssets() *btree.BTree {
	return a.assets
}

// ReplaceOrInsert for assets
func (a *MarginAccount) AssetUpdate(item UserAsset) {
	val := UserAsset(item)
	a.assets.ReplaceOrInsert(&val)
}

func NewMargin(client *binance.Client, symbols []string) (al *MarginAccount, err error) {
	marginAccount, err := client.NewGetMarginAccountService().Do(context.Background())
	if err != nil {
		return
	}
	al = &MarginAccount{
		BorrowEnabled:       marginAccount.BorrowEnabled,
		MarginLevel:         marginAccount.MarginLevel,
		TotalAssetOfBTC:     marginAccount.TotalAssetOfBTC,
		TotalLiabilityOfBTC: marginAccount.TotalLiabilityOfBTC,
		TotalNetAssetOfBTC:  marginAccount.TotalNetAssetOfBTC,
		TradeEnabled:        marginAccount.TradeEnabled,
		TransferEnabled:     marginAccount.TransferEnabled,
		assets:              btree.New(2),
		mu:                  sync.Mutex{},
		assetsRestriction:   make(map[string]bool), // Add the missing field "mapSymbols"
	}
	for _, asset := range symbols {
		al.assetsRestriction[asset] = true
	}

	for _, asset := range marginAccount.UserAssets {
		if _, exists := al.assetsRestriction[asset.Asset]; exists || len(al.assetsRestriction) == 0 {
			al.AssetUpdate(UserAsset(asset))
		}
	}
	return
}
