package account

import (
	"context"
	"errors"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	balances_types "github.com/fr0ster/go-trading-utils/types/balances"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Asset    futures.AccountAsset
	Position futures.AccountPosition
	Account  struct {
		client           *futures.Client
		account          *futures.Account
		accountAssets    *btree.BTree
		accountPositions *btree.BTree
		mu               sync.Mutex
		symbols          map[string]bool
		symbolsRestrict  []string
	}
)

func (a *Asset) Less(item btree.Item) bool {
	return a.Asset < item.(*Asset).Asset
}

func (a *Asset) Equal(item btree.Item) bool {
	return a.Asset == item.(*Asset).Asset
}

func (a *Position) Less(item btree.Item) bool {
	return a.Symbol < item.(*Position).Symbol
}

func (a *Position) Equal(item btree.Item) bool {
	return a.Symbol == item.(*Position).Symbol
}

func (a *Account) GetAsset(asset string) (float64, error) {
	item := a.accountAssets.Get(&Asset{Asset: asset})
	if item == nil {
		item = a.accountPositions.Get(&Position{Symbol: asset})
		if item == nil {
			return 0, errors.New("item not found")
		} else {
			symbolBalance := item.(*Position).MaintMargin
			return utils.ConvStrToFloat64(symbolBalance), nil
		}
	} else {
		symbolBalance := item.(*Asset).AvailableBalance
		return utils.ConvStrToFloat64(symbolBalance), nil
	}
}

func (a *Account) GetLockedAsset(asset string) (float64, error) {
	item := a.accountAssets.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, err := Binance2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.AvailableBalance), err
	}
}

func (a *Account) GetTotalAsset(asset string) (float64, error) {
	item := a.accountAssets.Get(&balances_types.BalanceItemType{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, err := Binance2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.WalletBalance), err
	}
}

func (a *Account) GetAssets() []*futures.AccountAsset {
	return a.account.Assets
}

func (a *Account) GetPositions() []*futures.AccountPosition {
	return a.account.Positions
}

func (a *Account) GetPositionRisk(symbol string) ([]*futures.PositionRisk, error) {
	risk, err := a.client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return risk, nil
}

// GetBalances implements account.AccountLimits.
func (a *Account) GetBalances() *btree.BTree {
	return a.accountAssets
}

func (a *Account) AssetsAscend(iterator func(item *balances_types.BalanceItemType) bool) {
	a.accountAssets.Ascend(func(i btree.Item) bool {
		return iterator(i.(*balances_types.BalanceItemType))
	})
}

func (a *Account) PositionsAscend(iterator func(item *balances_types.BalanceItemType) bool) {
	a.accountPositions.Ascend(func(i btree.Item) bool {
		return iterator(i.(*balances_types.BalanceItemType))
	})
}

func (a *Account) AssetsDescend(iterator func(item *balances_types.BalanceItemType) bool) {
	a.accountAssets.Descend(func(i btree.Item) bool {
		return iterator(i.(*balances_types.BalanceItemType))
	})
}

func (a *Account) PositionsDescend(iterator func(item *balances_types.BalanceItemType) bool) {
	a.accountPositions.Descend(func(i btree.Item) bool {
		return iterator(i.(*balances_types.BalanceItemType))
	})
}

func (a *Account) GetAssetsTree() *btree.BTree {
	return a.accountAssets
}

func (a *Account) GetPositionsTree() *btree.BTree {
	return a.accountPositions
}

func (a *Account) Update() error {
	var err error
	for _, symbol := range a.symbolsRestrict {
		a.symbols[symbol] = true
	}
	a.account, err = a.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return err
	}
	for _, asset := range a.account.Assets {
		if _, exists := a.symbols[asset.Asset]; exists || len(a.symbols) == 0 {
			val, err := Binance2AccountAsset(asset)
			if err != nil {
				continue
			}
			a.accountAssets.ReplaceOrInsert(val)
		}
	}
	for _, position := range a.account.Positions {
		if _, exists := a.symbols[position.Symbol]; exists || len(a.symbols) == 0 {
			val, err := Binance2AccountPosition(position)
			if err != nil {
				continue
			}
			a.accountPositions.ReplaceOrInsert(val)
		}
	}
	return nil
}

func New(client *futures.Client, degree int, symbols []string) (al *Account, err error) {
	al = &Account{
		client:           client,
		account:          nil,
		accountAssets:    btree.New(degree),
		accountPositions: btree.New(degree),
		mu:               sync.Mutex{},
		symbols:          make(map[string]bool),
		symbolsRestrict:  symbols,
	}
	err = al.Update()
	return
}

func Binance2AccountAsset(binanceAsset interface{}) (*Asset, error) {
	var asset Asset
	err := copier.Copy(&asset, binanceAsset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func Binance2AccountPosition(binancePosition interface{}) (*Position, error) {
	var position Position
	err := copier.Copy(&position, binancePosition)
	if err != nil {
		return nil, err
	}
	return &position, nil
}
