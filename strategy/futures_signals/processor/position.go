package processor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	pairs_types "github.com/fr0ster/go-trading-utils/types/pairs"
	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"
	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) getPositionRisk(times int) (risks []*futures.PositionRisk, err error) {
	if times == 0 {
		return
	}
	risks, err = pp.client.NewGetPositionRiskService().Symbol(pp.pairInfo.GetSymbol()).Do(context.Background())
	if err != nil {
		errApi, _ := utils.ParseAPIError(err)
		if errApi != nil && errApi.Code == -1021 {
			time.Sleep(3 * time.Second)
			return pp.getPositionRisk(times - 1)
		}
	}
	return
}

func (pp *PairProcessor) GetPositionRisk() (risks *futures.PositionRisk, err error) {
	risk, err := pp.getPositionRisk(3)
	if err != nil {
		return nil, err
	} else if len(risk) == 0 {
		return nil, fmt.Errorf("can't get position risk for symbol %s", pp.symbol.Symbol)
	} else {
		return risk[0], nil
	}
}

func (pp *PairProcessor) GetLiquidationDistance(price float64) (distance float64) {
	risk, _ := pp.GetPositionRisk()
	return math.Abs((price - utils.ConvStrToFloat64(risk.LiquidationPrice)) / utils.ConvStrToFloat64(risk.LiquidationPrice))
}

func (pp *PairProcessor) GetNotional() float64 {
	return pp.notional
}

func (pp *PairProcessor) GetLeverage() int {
	if pp.leverage == 0 {
		risk, _ := pp.GetPositionRisk()
		pp.leverage = int(utils.ConvStrToFloat64(risk.Leverage))
	}
	return pp.leverage
}

func (pp *PairProcessor) SetLeverage(leverage int) (res *futures.SymbolLeverage, err error) {
	return pp.client.NewChangeLeverageService().Symbol(pp.symbol.Symbol).Leverage(leverage).Do(context.Background())
}

func (pp *PairProcessor) GetCallbackRate() float64 {
	return pp.callbackRate
}

func (pp *PairProcessor) SetCallbackRate(callbackRate float64) {
	pp.callbackRate = callbackRate
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) GetMarginType() pairs_types.MarginType {
	if pp.marginType == "" {
		risk, _ := pp.GetPositionRisk()
		pp.marginType = pairs_types.MarginType(risk.MarginType)
	}
	return pp.marginType
}

// MarginTypeIsolated MarginType = "ISOLATED"
// MarginTypeCrossed  MarginType = "CROSSED"
func (pp *PairProcessor) SetMarginType(marginType pairs_types.MarginType) (err error) {
	return pp.client.
		NewChangeMarginTypeService().
		Symbol(pp.symbol.Symbol).
		MarginType(futures.MarginType(marginType)).
		Do(context.Background())
}

func (pp *PairProcessor) GetPositionMargin() (margin float64) {
	risk, err := pp.GetPositionRisk()
	if err != nil {
		return 0
	}
	margin = utils.ConvStrToFloat64(risk.IsolatedMargin) // Convert string to float64
	return
}

func (pp *PairProcessor) SetPositionMargin(amountMargin float64, typeMargin int) (err error) {
	return pp.client.NewUpdatePositionMarginService().
		Symbol(pp.symbol.Symbol).Type(typeMargin).
		Amount(utils.ConvFloat64ToStrDefault(amountMargin)).Do(context.Background())
}

func (pp *PairProcessor) GetSymbol() *symbol_types.FuturesSymbol {
	// Ініціалізуємо інформацію про пару
	return pp.pairInfo
}

func (pp *PairProcessor) GetFuturesSymbol() (*futures.Symbol, error) {
	return pp.symbol, nil
}

// Округлення ціни до StepSize знаків після коми
func (pp *PairProcessor) GetStepSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TickSize знаків після коми
func (pp *PairProcessor) GetTickSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))))
}

func (pp *PairProcessor) GetAccount() (account *futures.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetPair() string {
	return pp.symbol.Symbol
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

func (pp *PairProcessor) GetBaseBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.WalletBalance) // Convert string to float64
	return
}

func (pp *PairProcessor) GetTargetBalance() (balance float64, err error) {
	asset, err := pp.GetTargetAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	return
}

func (pp *PairProcessor) GetFreeBalance() (balance float64) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return 0
	}
	balance = utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	if balance > pp.limitOnPosition {
		balance = pp.limitOnPosition
	}
	return
}

func (pp *PairProcessor) GetLockedBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance) // Convert string to float64
	return
}
