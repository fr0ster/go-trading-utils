package processor

import (
	"context"
	"fmt"
	"math"

	"github.com/adshao/go-binance/v2/futures"
	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) GetAccount() (account *futures.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetPair() string {
	return pp.pairInfo.Symbol
}

func (pp *PairProcessor) GetSymbol() *symbol_types.SymbolInfo {
	// Ініціалізуємо інформацію про пару
	return pp.pairInfo
}

// func (pp *PairProcessor) GetFuturesSymbol() (*futures.Symbol, error) {
// 	return pp.symbol, nil
// }

// Округлення ціни до StepSize знаків після коми
func (pp *PairProcessor) GetStepSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(float64(pp.pairInfo.GetStepSize())))))
}

// Округлення ціни до TickSize знаків після коми
func (pp *PairProcessor) GetTickSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(float64(pp.pairInfo.GetTickSizeExp())))))
}

func (pp *PairProcessor) GetNotional() items_types.ValueType {
	return pp.notional
}

func (pp *PairProcessor) GetCallbackRate() items_types.PricePercentType {
	return pp.callbackRate
}

func (pp *PairProcessor) SetCallbackRate(callbackRate items_types.PricePercentType) {
	pp.callbackRate = callbackRate
}

func (pp *PairProcessor) GetBaseAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.baseSymbol {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.baseSymbol)
}

func (pp *PairProcessor) GetTargetAsset() (asset *futures.AccountAsset, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Assets {
		if asset.Asset == pp.targetSymbol {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.targetSymbol)
}

func (pp *PairProcessor) GetBaseBalance() (balance items_types.ValueType, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
	return
}

func (pp *PairProcessor) GetTargetBalance() (balance items_types.ValueType, err error) {
	asset, err := pp.GetTargetAsset()
	if err != nil {
		return
	}
	balance = items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
	return
}

func (pp *PairProcessor) GetFreeBalance() (balance items_types.ValueType) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return 0
	}
	balance = items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
	if balance > pp.limitOnPosition {
		balance = pp.limitOnPosition
	}
	return
}

func (pp *PairProcessor) GetLockedBalance() (balance items_types.ValueType, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance))
	return
}

func (pp *PairProcessor) GetCurrentPrice() (items_types.PriceType, error) {
	price, err := pp.client.NewListPricesService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price)), nil
}
