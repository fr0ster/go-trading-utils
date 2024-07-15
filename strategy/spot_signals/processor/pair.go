package processor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/adshao/go-binance/v2"

	symbol_types "github.com/fr0ster/go-trading-utils/types/symbol"

	utils "github.com/fr0ster/go-trading-utils/utils"
)

func (pp *PairProcessor) SetTimeOut(timeOut time.Duration) {
	pp.timeOut = timeOut
}

func (pp *PairProcessor) GetSymbol() *symbol_types.SpotSymbol {
	return pp.pairInfo
}

func (pp *PairProcessor) GetCurrentPrice() (float64, error) {
	price, err := pp.client.NewListPricesService().Symbol(pp.pairInfo.Symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func (pp *PairProcessor) GetPair() string {
	return pp.pairInfo.Symbol
}

func (pp *PairProcessor) GetAccount() (account *binance.Account, err error) {
	return pp.client.NewGetAccountService().Do(context.Background())
}

func (pp *PairProcessor) GetBaseAsset() (asset *binance.Balance, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Balances {
		if asset.Asset == pp.baseSymbol {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.baseSymbol)
}

func (pp *PairProcessor) GetTargetAsset() (asset *binance.Balance, err error) {
	account, err := pp.GetAccount()
	if err != nil {
		return
	}
	for _, asset := range account.Balances {
		if asset.Asset == pp.targetSymbol {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("can't find asset %s", pp.targetSymbol)
}

func (pp *PairProcessor) GetBaseBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
	return
}

func (pp *PairProcessor) GetTargetBalance() (balance float64, err error) {
	asset, err := pp.GetTargetAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
	return
}

func (pp *PairProcessor) GetFreeBalance() (balance float64) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return 0
	}
	balance = utils.ConvStrToFloat64(asset.Free) // Convert string to float64
	if balance > pp.limitOnPosition {
		balance = pp.limitOnPosition
	}
	return
}

func (pp *PairProcessor) GetLimitOnTransaction() (limit float64) {
	return pp.limitOnTransaction * pp.GetFreeBalance()
}

func (pp *PairProcessor) SetBounds(price float64) {
	pp.UpBound = price * (1 + pp.UpBoundPercent)
	pp.LowBound = price * (1 - pp.LowBoundPercent)
}

func (pp *PairProcessor) GetUpBound() float64 {
	return pp.UpBound
}

func (pp *PairProcessor) GetLowBound() float64 {
	return pp.LowBound
}

func (pp *PairProcessor) GetCallbackRate() float64 {
	return pp.callbackRate
}

func (pp *PairProcessor) SetCallbackRate(callbackRate float64) {
	pp.callbackRate = callbackRate
}

func (pp *PairProcessor) GetDeltaPrice() float64 {
	return pp.deltaPrice
}

func (pp *PairProcessor) SetDeltaPrice(deltaPrice float64) {
	pp.deltaPrice = deltaPrice
}

func (pp *PairProcessor) GetDeltaQuantity() float64 {
	return pp.deltaQuantity
}

func (pp *PairProcessor) GetLockedBalance() (balance float64, err error) {
	asset, err := pp.GetBaseAsset()
	if err != nil {
		return
	}
	balance = utils.ConvStrToFloat64(asset.Locked) // Convert string to float64
	return
}

// Округлення ціни до SteGSize знаків після коми
func (pp *PairProcessor) GetStepSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.LotSizeFilter().StepSize)))))
}

// Округлення ціни до TicGSize знаків після коми
func (pp *PairProcessor) GetTickSizeExp() int {
	return int(math.Abs(math.Round(math.Log10(utils.ConvStrToFloat64(pp.symbol.PriceFilter().TickSize)))))
}

func (pp *PairProcessor) GetNotional() float64 {
	return pp.notional
}
