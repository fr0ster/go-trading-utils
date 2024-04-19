package account

import (
	"context"
	"errors"
	"sync"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/google/btree"
	"github.com/jinzhu/copier"
)

type (
	Asset    futures.AccountAsset
	Position futures.AccountPosition
	Account  struct {
		client                      *futures.Client
		FeeTier                     int    `json:"feeTier"`
		CanTrade                    bool   `json:"canTrade"`
		CanDeposit                  bool   `json:"canDeposit"`
		CanWithdraw                 bool   `json:"canWithdraw"`
		UpdateTime                  int64  `json:"updateTime"`
		MultiAssetsMargin           bool   `json:"multiAssetsMargin"`
		TotalInitialMargin          string `json:"totalInitialMargin"`
		TotalMaintMargin            string `json:"totalMaintMargin"`
		TotalWalletBalance          string `json:"totalWalletBalance"`
		TotalUnrealizedProfit       string `json:"totalUnrealizedProfit"`
		TotalMarginBalance          string `json:"totalMarginBalance"`
		TotalPositionInitialMargin  string `json:"totalPositionInitialMargin"`
		TotalOpenOrderInitialMargin string `json:"totalOpenOrderInitialMargin"`
		TotalCrossWalletBalance     string `json:"totalCrossWalletBalance"`
		TotalCrossUnPnl             string `json:"totalCrossUnPnl"`
		AvailableBalance            string `json:"availableBalance"`
		MaxWithdrawAmount           string `json:"maxWithdrawAmount"`
		assets                      *btree.BTree
		positions                   *btree.BTree
		mu                          sync.Mutex
		assetsName                  map[string]bool
		symbolsName                 map[string]bool
		assetsRestrict              []string
		symbolsRestrict             []string
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

func (a *Account) Lock() {
	a.mu.Lock()
}

func (a *Account) Unlock() {
	a.mu.Unlock()
}

func (a *Account) GetFreeAsset(asset string) (float64, error) {
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		item = a.positions.Get(&Position{Symbol: asset})
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
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, err := Futures2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.AvailableBalance), err
	}
}

func (a *Account) GetTotalAsset(asset string) (float64, error) {
	item := a.assets.Get(&Asset{Asset: asset})
	if item == nil {
		return 0, errors.New("item not found")
	} else {
		symbolBalance, err := Futures2AccountAsset(item)
		return utils.ConvStrToFloat64(symbolBalance.WalletBalance), err
	}
}

func (a *Account) GetPositionRisk(symbol string) ([]*futures.PositionRisk, error) {
	risk, err := a.client.NewGetPositionRiskService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return risk, nil
}

// GetBalances implements account.AccountLimits.
func (a *Account) GetAssets() *btree.BTree {
	return a.assets
}

// GetPositions implements account.AccountLimits.
func (a *Account) GetPositions() *btree.BTree {
	return a.positions
}

func (a *Account) AssetsAscend(iterator func(item *Asset) bool) {
	a.assets.Ascend(func(i btree.Item) bool {
		return iterator(i.(*Asset))
	})
}

func (a *Account) PositionsAscend(iterator func(item *Position) bool) {
	a.positions.Ascend(func(i btree.Item) bool {
		return iterator(i.(*Position))
	})
}

func (a *Account) AssetsDescend(iterator func(item *Asset) bool) {
	a.assets.Descend(func(i btree.Item) bool {
		return iterator(i.(*Asset))
	})
}

func (a *Account) PositionsDescend(iterator func(item *Position) bool) {
	a.positions.Descend(func(i btree.Item) bool {
		return iterator(i.(*Position))
	})
}

// ReplaceOrInsert for Assets
func (a *Account) AssetUpdate(item *Asset) {
	a.assets.ReplaceOrInsert(item)
}

// ReplaceOrInsert for Positions
func (a *Account) PositionsUpdate(item *Position) {
	a.positions.ReplaceOrInsert(item)
}

func New(client *futures.Client, degree int, assets []string, symbols []string) (*Account, error) {
	accountIn, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	account := &Account{
		client:                      client,
		FeeTier:                     accountIn.FeeTier,
		CanTrade:                    accountIn.CanTrade,
		CanDeposit:                  accountIn.CanDeposit,
		CanWithdraw:                 accountIn.CanWithdraw,
		UpdateTime:                  accountIn.UpdateTime,
		MultiAssetsMargin:           accountIn.MultiAssetsMargin,
		TotalInitialMargin:          accountIn.TotalInitialMargin,
		TotalMaintMargin:            accountIn.TotalMaintMargin,
		TotalWalletBalance:          accountIn.TotalWalletBalance,
		TotalUnrealizedProfit:       accountIn.TotalUnrealizedProfit,
		TotalMarginBalance:          accountIn.TotalMarginBalance,
		TotalPositionInitialMargin:  accountIn.TotalPositionInitialMargin,
		TotalOpenOrderInitialMargin: accountIn.TotalOpenOrderInitialMargin,
		TotalCrossWalletBalance:     accountIn.TotalCrossWalletBalance,
		TotalCrossUnPnl:             accountIn.TotalCrossUnPnl,
		AvailableBalance:            accountIn.AvailableBalance,
		MaxWithdrawAmount:           accountIn.MaxWithdrawAmount,
		assets:                      btree.New(degree),
		positions:                   btree.New(degree),
		mu:                          sync.Mutex{},
		assetsName:                  make(map[string]bool),
		symbolsName:                 make(map[string]bool),
		assetsRestrict:              assets,
		symbolsRestrict:             symbols,
	}
	for _, asset := range account.assetsRestrict {
		account.assetsName[asset] = true
	}
	for _, symbol := range account.symbolsRestrict {
		account.symbolsName[symbol] = true
	}
	for _, asset := range accountIn.Assets {
		if _, exists := account.assetsName[asset.Asset]; exists || len(account.assetsName) == 0 {
			val, err := Futures2AccountAsset(asset)
			if err != nil {
				continue
			}
			account.assets.ReplaceOrInsert(val)
		}
	}
	for _, position := range accountIn.Positions {
		if _, exists := account.symbolsName[position.Symbol]; exists || len(account.symbolsName) == 0 {
			val, err := Futures2AccountPosition(position)
			if err != nil {
				continue
			}
			account.positions.ReplaceOrInsert(val)
		}
	}
	return account, nil
}

func Futures2AccountAsset(binanceAsset interface{}) (*Asset, error) {
	var asset Asset
	err := copier.Copy(&asset, binanceAsset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func Futures2AccountPosition(binancePosition interface{}) (*Position, error) {
	var position Position
	err := copier.Copy(&position, binancePosition)
	if err != nil {
		return nil, err
	}
	return &position, nil
}
